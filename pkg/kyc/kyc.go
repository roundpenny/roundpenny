package kyc

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type Client struct {
	apiToken   string
	mock       bool
	httpClient *http.Client
}

type Applicant struct {
	ID        string `json:"id,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
}

type CheckResult struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Result        string `json:"result,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	DocumentID    string `json:"document_id,omitempty"`
}

type onfidoApplicantRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
}

type onfidoApplicantResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type onfidoCheckRequest struct {
	ApplicantID  string   `json:"applicant_id"`
	ReportNames  []string `json:"report_names"`
	DocumentIDs  []string `json:"document_ids,omitempty"`
}

type onfidoCheckResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Result    string `json:"result,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

var (
	ErrMissingAPIToken = errors.New("ONFIDO_API_TOKEN not set")
	ErrCheckFailed     = errors.New("KYC check failed")
	ErrAPICallFailed   = errors.New("Onfido API call failed")
)

func NewClient() *Client {
	token := os.Getenv("ONFIDO_API_TOKEN")
	mock := token == ""

	if mock {
		slog.Warn("ONFIDO_API_TOKEN not set, using mock KYC mode")
	}

	return &Client{
		apiToken:   token,
		mock:       mock,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) CreateApplicant(firstName, lastName, email string) (*Applicant, error) {
	if c.mock {
		return c.mockCreateApplicant(firstName, lastName, email)
	}

	body := onfidoApplicantRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
	payload, _ := json.Marshal(body)

	resp, err := c.doRequest("POST", "https://api.onfido.com/v3.6/applicants", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result onfidoApplicantResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("%w: decode error: %v", ErrAPICallFailed, err)
	}

	return &Applicant{
		ID:        result.ID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
	}, nil
}

func (c *Client) UploadDocument(applicantID, filePath, docType string) (string, error) {
	if c.mock {
		return c.mockUploadDocument(applicantID)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	w.WriteField("applicant_id", applicantID)
	w.WriteField("type", docType)
	part, _ := w.CreateFormFile("file", filePath)
	io.Copy(part, file)
	w.Close()

	req, err := http.NewRequest("POST", "https://api.onfido.com/v3.6/documents", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Token token="+c.apiToken)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: decode error: %v", ErrAPICallFailed, err)
	}

	return result.ID, nil
}

func (c *Client) CreateCheck(applicantID string, documentIDs []string) (*CheckResult, error) {
	if c.mock {
		return c.mockCreateCheck(applicantID)
	}

	body := onfidoCheckRequest{
		ApplicantID: applicantID,
		ReportNames: []string{"identity_enhanced", "document"},
		DocumentIDs: documentIDs,
	}
	payload, _ := json.Marshal(body)

	resp, err := c.doRequest("POST", "https://api.onfido.com/v3.6/checks", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result onfidoCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("%w: decode error: %v", ErrAPICallFailed, err)
	}

	checkResult := &CheckResult{
		ID:        result.ID,
		Status:    result.Status,
		Result:    result.Result,
		CreatedAt: result.CreatedAt,
	}

	return checkResult, nil
}

func (c *Client) doRequest(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token token="+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("%w: status %d: %s", ErrAPICallFailed, resp.StatusCode, string(respBody))
	}

	return resp, nil
}

func (c *Client) mockCreateApplicant(firstName, lastName, email string) (*Applicant, error) {
	slog.Info("mock KYC: create applicant", "name", firstName+" "+lastName)
	return &Applicant{
		ID:        "app_mock_" + randHex(8),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}, nil
}

func (c *Client) mockUploadDocument(applicantID string) (string, error) {
	slog.Info("mock KYC: upload document", "applicant", applicantID)
	return "doc_mock_" + randHex(8), nil
}

func (c *Client) mockCreateCheck(applicantID string) (*CheckResult, error) {
	slog.Info("mock KYC: create check", "applicant", applicantID)
	return &CheckResult{
		ID:        "chk_mock_" + randHex(8),
		Status:    "complete",
		Result:    "clear",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://kong:8000';

const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const profileDuration = new Trend('profile_duration');
const merchantDuration = new Trend('merchant_duration');

export const options = {
  stages: [
    { duration: '10s', target: 10 },
    { duration: '20s', target: 20 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    errors: ['rate<0.01'],
    http_req_duration: ['p(95)<2000'],
    login_duration: ['p(95)<2000'],
  },
};

function getEmail(vu) {
  return `loadtest.${vu}@test.com`;
}

export function setup() {
  // Pre-register all test users
  for (let i = 1; i <= 20; i++) {
    const email = getEmail(i);
    http.post(`${BASE_URL}/v1/auth/register`, JSON.stringify({
      email: email,
      password: 'password123',
      full_name: `Load Test User ${i}`,
    }), { headers: { 'Content-Type': 'application/json' } });
  }
  return {};
}

export default function () {
  const email = getEmail(__VU);

  group('mixed', function () {
    const loginRes = http.post(`${BASE_URL}/v1/auth/login`, JSON.stringify({
      email: email,
      password: 'password123',
    }), { headers: { 'Content-Type': 'application/json' } });

    loginDuration.add(loginRes.timings.duration);
    errorRate.add(loginRes.status >= 400);
    check(loginRes, { 'login status 200': (r) => r.status === 200 });

    const token = loginRes.json('access_token');
    if (token) {
      const profileRes = http.get(`${BASE_URL}/v1/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      profileDuration.add(profileRes.timings.duration);
      errorRate.add(profileRes.status >= 400);
      check(profileRes, { 'profile status 200': (r) => r.status === 200 });

      const merchRes = http.post(`${BASE_URL}/v1/merchants`, JSON.stringify({
        name: `Merchant ${__VU}`,
        email: `merchant.${__VU}@test.com`,
        description: 'Load test merchant',
      }), {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
      });
      merchantDuration.add(merchRes.timings.duration);
      errorRate.add(merchRes.status >= 400);
      check(merchRes, { 'merchant create status 201': (r) => r.status === 201 });
    }
  });

  sleep(1);
}

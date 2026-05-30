// Copyright (c) 2026 RoundPenny. All rights reserved.

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const registerFailRate = new Rate('register_failures');
const loginFailRate = new Rate('login_failures');
const paymentFailRate = new Rate('payment_failures');
const latencyTrend = new Trend('end_to_end_latency');
const throughputCounter = new Counter('requests_completed');

// Test configuration
export let options = {
    stages: [
        { duration: '30s', target: 20 },   // Ramp up to 20 VU
        { duration: '1m', target: 50 },     // Ramp to 50 VU
        { duration: '2m', target: 100 },    // Ramp to 100 VU
        { duration: '2m', target: 100 },    // Stay at 100 VU
        { duration: '30s', target: 50 },    // Ramp down
        { duration: '30s', target: 0 },     // Cool down
    ],
    thresholds: {
        http_req_duration: ['p(95)<2000', 'p(99)<5000'],
        http_req_failed: ['rate<0.01'],
        register_failures: ['rate<0.05'],
        login_failures: ['rate<0.05'],
        payment_failures: ['rate<0.05'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';
const ADMIN_EMAIL = 'admin@roundpenny.com';
const ADMIN_PASSWORD = 'admin123';

// Users pool
const users = [];
for (let i = 0; i < 50; i++) {
    users.push({
        email: `loadtest${i}@test.com`,
        password: 'TestPass123!',
        name: `Load Tester ${i}`,
    });
}

function registerUser(user) {
    const res = http.post(`${BASE_URL}/v1/auth/register`, JSON.stringify({
        email: user.email,
        password: user.password,
        name: user.name,
    }), { headers: { 'Content-Type': 'application/json' } });
    registerFailRate.add(res.status !== 201 && res.status !== 409);
    return res.status === 201;
}

function loginUser(user) {
    const res = http.post(`${BASE_URL}/v1/auth/login`, JSON.stringify({
        email: user.email,
        password: user.password,
    }), { headers: { 'Content-Type': 'application/json' } });
    loginFailRate.add(res.status !== 200);
    if (res.status === 200) {
        const body = JSON.parse(res.body);
        return body.access_token;
    }
    return null;
}

export function setup() {
    // Pre-register users
    const registered = [];
    for (const user of users) {
        const res = http.post(`${BASE_URL}/v1/auth/register`, JSON.stringify({
            email: user.email,
            password: user.password,
            name: user.name,
        }), { headers: { 'Content-Type': 'application/json' } });
        if (res.status === 201 || res.status === 409) {
            registered.push(user);
        }
    }
    return { users: registered };
}

export default function(data) {
    group('auth_flow', function() {
        const user = data.users[Math.floor(Math.random() * data.users.length)];
        const token = loginUser(user);
        if (token) {
            check(token, { 'logged in': (t) => t !== null });
            
            // Get profile
            const profileRes = http.get(`${BASE_URL}/v1/users/me`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            check(profileRes, { 'profile ok': (r) => r.status === 200 });
            
            // List merchants
            const merchantsRes = http.get(`${BASE_URL}/v1/merchants`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            
            // Create payment (if idempotency key provided)
            const paymentRes = http.post(`${BASE_URL}/v1/payments`, JSON.stringify({
                amount: Math.random() * 100 + 1,
                currency: 'USD',
                payment_method: 'card',
            }), { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}`, 'Idempotency-Key': `loadtest-${__VU}-${__ITER}` } });
            paymentFailRate.add(paymentRes.status !== 201 && paymentRes.status !== 409);
            
            // Get wallets
            const walletsRes = http.get(`${BASE_URL}/v1/wallets`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            
            // Get transactions
            const txRes = http.get(`${BASE_URL}/v1/transactions?page=1&page_size=10`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            
            latencyTrend.add(Date.now());
            throughputCounter.add(1);
        }
        sleep(Math.random() * 3 + 1);
    });
    
    // Admin flow (every 10th iteration)
    if (__ITER % 10 === 0) {
        group('admin_flow', function() {
            const loginRes = http.post(`${BASE_URL}/v1/auth/login`, JSON.stringify({
                email: ADMIN_EMAIL,
                password: ADMIN_PASSWORD,
            }), { headers: { 'Content-Type': 'application/json' } });
            
            if (loginRes.status === 200) {
                const adminToken = JSON.parse(loginRes.body).access_token;
                
                http.get(`${BASE_URL}/v1/admin/dashboard`, {
                    headers: { 'Authorization': `Bearer ${adminToken}` }
                });
                http.get(`${BASE_URL}/v1/admin/users?page=1&page_size=10`, {
                    headers: { 'Authorization': `Bearer ${adminToken}` }
                });
                http.get(`${BASE_URL}/v1/admin/transactions?page=1&page_size=10`, {
                    headers: { 'Authorization': `Bearer ${adminToken}` }
                });
            }
        });
    }
}

export function teardown(data) {
    // Cleanup is handled by TTL in the system
    console.log(`Load test completed. Total users: ${data.users.length}`);
}

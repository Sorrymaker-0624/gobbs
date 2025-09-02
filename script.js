import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';

export const options = {
    scenarios: {
        read_detail_heavy: {
            executor: 'constant-vus',
            vus: 1200,
            duration: '45s',
            exec: 'readDetail',
            tags: { scenario: 'read' },
        },
        write_like_heavy: {
            executor: 'ramping-arrival-rate',
            startRate: 500,
            timeUnit: '1s',
            preAllocatedVUs: 600,
            maxVUs: 2000,
            stages: [
                { target: 1500, duration: '20s' },
                { target: 2500, duration: '20s' },
                { target: 0, duration: '5s' },
            ],
            exec: 'likeWrite',
            tags: { scenario: 'write' },
        },
    },
    thresholds: {
        'http_req_duration{scenario:read}': ['p(95)<60'],
        'http_req_duration{scenario:write}': ['p(95)<120'],
        'http_req_failed': ['rate<0.01'],
    },
};

const BASE = 'http://app:8080/api/v1';

function registerIfNeeded(username, password) {
    const payload = { username, password, confirm_password: password, email: `${username}@ex.com` };
    const params = { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } };
    const res = http.post(`${BASE}/register`, toForm(payload), params);
    return res.status === 200 || res.status === 409;
}

function login(username, password) {
    const params = { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } };
    const res = http.post(`${BASE}/login`, toForm({ username, password }), params);
    check(res, { 'login 200': (r) => r.status === 200 && r.json('session_id') });
    return res.json('session_id');
}

function createPost(token, title, content, community_id = 1) {
    const params = {
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            Authorization: `Bearer ${token}`,
        },
    };
    const res = http.post(`${BASE}/posts`, toForm({ title, content, community_id }), params);
    check(res, { 'create post ok': (r) => r.status === 200 && r.json('post_id') });
    return res.json('post_id');
}

function preloadCache(postId) {
    const res = http.get(`${BASE}/posts/${postId}`);
    check(res, { 'preload 200': (r) => r.status === 200 });
}

function toForm(obj) {
    return Object.keys(obj)
        .map((k) => `${encodeURIComponent(k)}=${encodeURIComponent(obj[k])}`)
        .join('&');
}

const setupData = new SharedArray('setup', function () {
    const username = `u_${__ENV.USER_SUFFIX || 'perf'}`;
    const password = 'P@ssw0rd!';

    registerIfNeeded(username, password);
    const token = login(username, password);

    // 预创建 N 篇帖子，返回第一个作为热点帖用于读压测
    const hotspotId = createPost(token, 'Hot Post', 'This is a hot post for cache.', 1);
    // 额外创建其他帖子避免唯一性束缚，给写压测用（点赞集合不要求存在于 DB）
    const writeIds = [];
    for (let i = 0; i < 5; i++) {
        writeIds.push(createPost(token, `Like Target ${i}`, 'for like set pressure', 1));
    }

    // 预热热点帖缓存
    for (let i = 0; i < 5; i++) preloadCache(hotspotId);

    return [{ token, hotspotId, writeIds }];
});

export function setup() {
    return setupData[0];
}

export function readDetail(data) {
    const res = http.get(`${BASE}/posts/${data.hotspotId}`);
    check(res, {
        'detail 200': (r) => r.status === 200,
        'detail fast': (r) => r.timings.duration < 80, // 目标靠近 35ms
    });
    sleep(0.2);
}

export function likeWrite(data) {
    // 多用户令牌共享一个 session 以提升点赞吞吐（服务端用 Redis Set 去重）
    const token = data.token;
    const headers = { Authorization: `Bearer ${token}` };

    const target = data.writeIds[exec.vu.idInTest % data.writeIds.length];
    const res = http.post(`${BASE}/posts/${target}/like`, null, { headers });
    check(res, { 'like 200': (r) => r.status === 200 });
    sleep(0.1);
}

export function teardown(data) {
    // 输出用于观测的热点帖子一次
    const res = http.get(`${BASE}/posts/${data.hotspotId}`);
    check(res, { 'final read ok': (r) => r.status === 200 });
}
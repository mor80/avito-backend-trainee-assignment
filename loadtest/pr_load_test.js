import { check } from 'k6';
import http from 'k6/http';

export const options = {
    scenarios: {
        create_pr: {
            executor: 'constant-arrival-rate',
            rate: 20,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 10,
            maxVUs: 50,
        },
    },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    const prId = `pr-${__VU}-${__ITER}`;

    const res = http.post(
        `${BASE_URL}/pullRequest/create`,
        JSON.stringify({
            pull_request_id: prId,
            pull_request_name: 'load-pr',
            author_id: 'u1',
        }),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    check(res, {
        'status is 201 or 409': (r) => r.status === 201 || r.status === 409,
    });
}

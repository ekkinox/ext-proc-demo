import http from 'k6/http';
import { check, fail } from 'k6';

export function setup() {
    return {
        globalRandom: Date.now() + Math.random(0) * 1000000000
    };
}

export default function (data) {

    let reqRandom = Date.now() + Math.random(0) * 1000000000;

    const response = http.post(
        'http://localhost:10000',
        {
            csrf: reqRandom
        },
        {
            headers: {
                'X-Cache': data.globalRandom
            }
        }
    );

    const checkOutput = check(response, {
        'expected 200': (resp) => resp.status == 200,
        'expected csrf': (resp) => resp.headers['X-Extracted-Csrf'] == reqRandom,
        'expected cache': (resp) => resp.headers['X-Extracted-Cache'] == data.globalRandom,
    });

    if (!checkOutput) {
        fail('unexpected response');
    }
}

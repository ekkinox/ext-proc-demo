import http from 'k6/http';
import { check, fail } from 'k6';

export default function () {

    let random = Date.now() + Math.random(0) * 1000000000;

    const response = http.post('http://localhost:10000', { csrf: random });

    const checkOutput = check(response, {
        'expected 200': (resp) => resp.status == 200,
        'expected csrf': (resp) => resp.headers['X-Extracted-Csrf'] == random,
    });

    if (!checkOutput) {
        fail('unexpected response');
    }
}

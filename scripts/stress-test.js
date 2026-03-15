import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    stages: [
        { duration: '30s', target: 500 },
        { duration: '1m', target: 2000 },
        { duration: '30s', target: 0 },
    ],
};

const siteIDs = [
    "fe1988f2-53b8-4505-90e1-29da5b8eeace", // OK
    "b2744901-013f-4cf4-88da-1ea4d53d5248", // OK
    "saas-app-main"                         // Not valid
];

const eventTypes = ["page_view", "click", "scroll", "form_submit"];

export default function () {
    const randomIp = `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`;

    const payload = JSON.stringify({
        event_id: uuidv4(),
        site_id: siteIDs[Math.floor(Math.random() * siteIDs.length)],
        type: eventTypes[Math.floor(Math.random() * eventTypes.length)],
        user_id: `user-${Math.floor(Math.random() * 1000)}`,
        timestamp: new Date().toISOString(),
        properties: { "source": "k6-test" }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Forwarded-For': randomIp,
        },
    };

    const res = http.post('http://gateway:8080/track', payload, params);

    check(res, { 'status is 202': (r) => r.status === 202 });

}
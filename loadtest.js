import http from "k6/http";
import { check } from "k6";
import { randomItem } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";

const baseUrl = `${__ENV.BASE_URL}/api/v1`;
const apiKey = __ENV.API_KEY;
const projectId = __ENV.PROJECT_ID;
const endpointId = __ENV.ENDPOINT_ID;

const names = ["John", "Jane", "Bert", "Ed"];
const emails = [
    "John@gmail.com",
    "Jane@amazon.com",
    "Bert@yahoo.com",
    "Ed@hotmail.com",
];

export const generateEventPayload = (endpointId) => ({
		endpoint_id: endpointId,
		event_type: "benchmark.event",
		custom_headers: {
				'X-Benchmark-Timestamp': Math.floor(Date.now() / 1000).toString()
		},
    data: {
        player_name: randomItem(names),
        email: randomItem(emails),
    },
});

export let options = {
		vus: 100,
		duration: '20s',
    noConnectionReuse: true,
    thresholds: {
        http_req_duration: ["p(99)<6000"], // 99% of requests must complete below 6s
    },
};

export default function () {
    let eventBody = JSON.stringify(generateEventPayload(endpointId));
    const createEventUrl = `${baseUrl}/projects/${projectId}/events`;

		let res = http.post(createEventUrl, eventBody, {
			headers: { 
				"Content-Type": "application/json",
				"Authorization": `Bearer ${apiKey}`,
			},
		});

		check(res, {
			'response code was 200': (res) => res.status == 200,
		});
}

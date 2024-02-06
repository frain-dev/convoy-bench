import * as globals from './globals.js';
import http from "k6/http";
import { check } from "k6";

const apiKey = __ENV.API_KEY;
const projectId = __ENV.PROJECT_ID;

export let options = globals.options
export default function () {
	let eventBody = JSON.stringify(globals.generate1kbPayload(globals.endpointId));
	const createEventUrl = `${globals.baseUrl}/projects/${projectId}/events`;

	let res = http.post(createEventUrl, eventBody, {
		timeout: "30s",
		headers: {
			"Content-Type": "application/json",
			"Authorization": `Bearer ${apiKey}`,
		},
	});

	check(res, {
		'response code was 201': (res) => res.status === 201,
		'response code was 200': (res) => res.status === 200,
		'response code was 4xx': (res) => res.status >= 400 && res.status < 500,
		'response code was 5xx': (res) => res.status >= 500,
	});
}

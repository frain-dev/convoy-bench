import * as globals from './globals.js';
import http from "k6/http";
import { check } from "k6";

export let options = globals.options


export default function () {
	let eventBody = JSON.stringify(globals.generate1kbPayload(globals.endpointId));

	let res = http.post(globals.ingestUrl, eventBody, {
		timeout: "30s",
		headers: {
			"Content-Type": "application/json",
			'X-Benchmark-Timestamp': Math.floor(Date.now() / 1000).toString()
		},
	});

	check(res, {
		'response code was 201': (res) => res.status === 201,
		'response code was 200': (res) => res.status === 200,
		'response code was 4xx': (res) => res.status >= 400 && res.status < 500,
		'response code was 5xx': (res) => res.status >= 500,
	});
}

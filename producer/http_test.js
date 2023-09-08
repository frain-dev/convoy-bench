import * as globals from './globals.js';
import http from "k6/http";
import { check } from "k6";

const apiKey = __ENV.API_KEY;
const projectId = __ENV.PROJECT_ID;

export let options = globals.options
export default function () {
	let eventBody = JSON.stringify(globals.generateSmallPayload(globals.endpointId));
  const createEventUrl = `${globals.baseUrl}/projects/${projectId}/events`;

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

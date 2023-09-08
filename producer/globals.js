import { randomItem } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";

const names = [
	"john",
	"jane", 
	"bert", 
	"ed",
]

const emails = [
	"john@gmail.com",
	"jane@amazon.com",
	"bert@yahoo.com",
	"ed@hotmail.com",
]

// Export global variables
export const baseUrl = `${__ENV.BASE_URL}/api/v1`;
export const endpointId = __ENV.ENDPOINT_ID;

// Export methods
export const generateSmallPayload = (endpointId) => ({
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

// Export k6 options
export const options = {
	vus: 20,
	duration: '15s',
	noConnectionReuse: true,
	thresholds: {
		http_req_duration: ["p(99)<6000"], // 99% of requests must complete below 6s.
	},
};

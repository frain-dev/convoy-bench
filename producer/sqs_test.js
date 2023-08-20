import * as globals from './globals.js';
import sqs from 'k6/x/sqs';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const client = sqs.newClient();
const queueUrl = __ENV.QUEUE_URL;

export default function () {
	const params = {
  	QueueUrl: queueUrl,
		MessageGroupId: 'default',
		MessageDeduplicationId: uuidv4(),
  	MessageBody: JSON.stringify(globals.generateSmallPayload(globals.endpointId)),
  };

	sqs.send(client,params)
}


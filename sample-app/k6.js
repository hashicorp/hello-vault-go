import http from 'k6/http';
import { sleep } from 'k6';

export default function () {
  http.post('http://localhost:8080/payments');
  sleep(1);
}

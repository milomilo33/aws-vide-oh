import axios from 'axios';

const apiBaseUrl = process.env.VUE_APP_REST_API_BASE_URL;
const apiKey = process.env.VUE_APP_API_KEY;

const apiClient = axios.create({
  baseURL: apiBaseUrl,
  headers: {
    'Content-Type': 'application/json',
    'x-api-key': apiKey
  },
});

export default apiClient;
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
});

// ---------- Tasks ----------
export const getTasks = async () => {
  const { data } = await api.get('/tasks');
  return data; // { tasks: [...] }
};

export const addTask = async (task) => {
  // task = { name, scheduled_time: "...", duration_sec }
  const { data } = await api.post('/tasks', task);
  return data;
};

export const markTaskAsDone = async (taskId) => {
  const { data } = await api.patch(`/tasks/${taskId}`);
  return data;
};

export const deleteTask = async (taskId) => {
  const { data } = await api.delete(`/tasks/${taskId}`);
  return data;
};

// ---------- Metrics (note the backend path is /metrices) ----------
export const getMetricsAll = async () => {
  const { data } = await api.get('/metrices', { params: { algo: 'all' } });
  return data; // { results: [...] }
};

export const getMetricsByAlgo = async (algo = 'fcfs') => {
  const { data } = await api.get('/metrices', { params: { algo } });
  return data; // single-algo shape
};

// Optional: unified error logging
api.interceptors.response.use(
  (res) => res,
  (err) => {
    console.error('API error:', err?.response?.data || err.message);
    return Promise.reject(err);
  }
);

export default api;
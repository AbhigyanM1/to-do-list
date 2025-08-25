import React, { useState } from 'react';
import { addTask } from '../services/api';

function AddTaskForm({ onTaskAdded }) {
  const [name, setName] = useState('');
  const [scheduledTime, setScheduledTime] = useState('');
  const [duration, setDuration] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name || !scheduledTime || !duration) {
      alert('Please fill all fields');
      return;
    }

    try {
      const newTask = await addTask({
        name,
        scheduled_time: scheduledTime,
        duration_sec: parseInt(duration, 10),
      });
      onTaskAdded?.(newTask);
      setName('');
      setScheduledTime('');
      setDuration('');
    } catch (err) {
      console.error('Failed to add task:', err);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ marginBottom: '20px' }}>
      <input
        type="text"
        placeholder="Task name"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <input
        type="datetime-local"
        value={scheduledTime}
        onChange={(e) => setScheduledTime(e.target.value)}
      />
      <input
        type="number"
        placeholder="Duration (sec)"
        value={duration}
        onChange={(e) => setDuration(e.target.value)}
      />
      <button type="submit">Add Task</button>
    </form>
  );
}

export default AddTaskForm;
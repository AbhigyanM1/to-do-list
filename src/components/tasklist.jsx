import React, { useEffect, useState, forwardRef, useImperativeHandle, useCallback } from 'react';
import Task from './task';
import { getTasks } from '../services/api';
const TaskList = forwardRef(({ onTaskUpdate }, ref) => {
  const [tasks, setTasks] = useState([]);

  const refreshTasks = useCallback(async () => {
    try {
      const response = await getTasks();
      if (response && Array.isArray(response.tasks)) {
        setTasks(response.tasks);
      } else {
        console.warn('Unexpected /tasks payload:', response);
      }
    } catch (error) {
      console.error('Failed to fetch tasks:', error);
    }
  }, []);

  const addTaskOptimistic = (task) => {
    if (!task) return;
    setTasks((prev) => [task, ...prev]);
  };

  // ➜ NEW: remove a task immediately after delete (prevents leftover strike-through)
  const handleTaskDeleted = (id) => {
    setTasks((prev) => prev.filter(t => t.id !== id));
    onTaskUpdate?.(); // optional: refresh graphs
  };

  useImperativeHandle(ref, () => ({
    refreshTasks,
    addTaskOptimistic,
  }), [refreshTasks]);

  useEffect(() => {
    refreshTasks();
  }, [refreshTasks]);

  return (
    <div>
      <h2>Task List</h2>
      {tasks.length === 0 ? (
        <p>No tasks found.</p>
      ) : (
        <ul>
          {tasks.map((task) => (
            <Task
              key={task.id}
              task={task}
              onTaskUpdated={onTaskUpdate}     // for Done
              onTaskDeleted={handleTaskDeleted} // for Delete
            />
          ))}
        </ul>
      )}
    </div>
  );
});

export default TaskList;
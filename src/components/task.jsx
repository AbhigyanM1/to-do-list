import React, { useState } from 'react';
import { markTaskAsDone, deleteTask } from '../services/api';

const Task = ({ task, refreshTasks, onTaskUpdate }) => {
  const [removed, setRemoved] = useState(false);
  const notify = () => onTaskUpdate && onTaskUpdate();

  const parseLocal = (s) => {
    if (!s) return null;
    const normalized = s.includes('T') ? s : s.replace(' ', 'T');
    const d = new Date(normalized);
    return isNaN(d.getTime()) ? null : d;
  };

  const schedAt = parseLocal(task.scheduled_time);
  const startAt = parseLocal(task.start_time);
  const endAt   = parseLocal(task.end_time);
  const now = new Date();

  const isDone      = !!endAt || task.done === 1;
  const isRunning   = !!startAt && !endAt;
  const isScheduled = !isRunning && !isDone;
  const isOverdue   = isScheduled && schedAt && now > schedAt;

  const statusLabel = isDone ? 'Done' : isRunning ? 'Running' : 'Scheduled';

  if (removed) return null;

  const taskStyle = {
    color: isOverdue ? 'grey' : 'black',
    opacity: isOverdue ? 0.6 : 1,
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    padding: '6px 0',
  };

  const badgeStyle = (bg) => ({
    fontSize: 12,
    padding: '2px 6px',
    borderRadius: 6,
    background: bg,
    color: '#fff',
  });

  const handleMarkDone = async () => {
    try {
      await markTaskAsDone(task.id);
      await refreshTasks?.();
      notify();
    } catch (error) {
      console.error('Failed to mark task as done:', error);
      alert('Failed to mark as done.');
    }
  };

  const handleDelete = async () => {
    try {
      setRemoved(true);
      await deleteTask(task.id);
      await refreshTasks?.();
      notify();
    } catch (error) {
      setRemoved(false);
      console.error('Failed to delete task:', error);
      alert('Failed to delete.');
    }
  };

  return (
    <li style={taskStyle}>
      {/* Strike-through only for the task name */}
      <span style={{ fontWeight: 600, textDecoration: isDone ? 'line-through' : 'none' }}>
        {task.name}
      </span>
      <span>•</span>
      <span title="Scheduled time">{task.scheduled_time}</span>

      {typeof task.duration_sec === 'number' && task.duration_sec > 0 && (
        <>
          <span>•</span>
          <span title="Planned run time">{task.duration_sec}s</span>
        </>
      )}

      <span>•</span>
      <span
        style={
          isDone
            ? badgeStyle('#16a34a')
            : isRunning
            ? badgeStyle('#f59e0b')
            : isOverdue
            ? badgeStyle('#6b7280')
            : badgeStyle('#3b82f6')
        }
      >
        {statusLabel}
      </span>

      <div style={{ marginLeft: 'auto', display: 'flex', gap: 8 }}>
        <button onClick={handleMarkDone} disabled={isDone}>
          Mark as Done
        </button>
        <button onClick={handleDelete}>Delete</button>
      </div>
    </li>
  );
};

export default Task;
import React, { useRef, useCallback, useEffect } from 'react';
import AddTaskForm from './components/addtaskform';
import TaskList from './components/tasklist';
import MetricsGraph from './components/metricsgraph';

function App() {
  const taskListRef = useRef(null);
  const graphRef = useRef(null);

  const refreshTasks = useCallback(() => {
    taskListRef.current?.refreshTasks();
  }, []);

  const refreshGraph = useCallback(() => {
    graphRef.current?.fetchMetrics?.();
  }, []);

  // Called after add / done / delete
  const handleTaskChange = useCallback((createdTask) => {
    // 1) optimistic insert if we just created one
    if (createdTask) {
      taskListRef.current?.addTaskOptimistic?.(createdTask);
    }
    // 2) reconcile from server
    taskListRef.current?.refreshTasks?.();
    // 3) then refresh graph (after tasks fetch kicks off)
    graphRef.current?.fetchMetrics?.();
  }, []);

  // Initial load: pull tasks + metrics once
  useEffect(() => {
    refreshTasks();
    refreshGraph();
  }, [refreshTasks, refreshGraph]);

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: 16 }}>
      <h1>Student To-Do List</h1>

      <AddTaskForm onTaskAdded={handleTaskChange} />

      <TaskList
        ref={taskListRef}
        onTaskUpdate={handleTaskChange}  // fires after done/delete or periodic refresh
      />

      {/* MetricsGraph should expose ref method: fetchMetrics() */}
      <MetricsGraph ref={graphRef} />
    </div>
  );
}

export default App;
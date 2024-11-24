import React, { useEffect, useState } from 'react';
import axios from 'axios';

function App() {
  const [aiServiceData, setAiServiceData] = useState('');
  const [postgresData, setPostgresData] = useState('');

  useEffect(() => {
    // Fetch data from core-service APIs
    axios.get('http://localhost:7012/api/data-from-ai-service')
      .then(response => setAiServiceData(response.data))
      .catch(error => console.error("Error fetching data from AI service:", error));

    axios.get('http://localhost:7012/api/data-from-postgres')
      .then(response => setPostgresData(response.data))
      .catch(error => console.error("Error fetching data from PostgreSQL:", error));
  }, []);

  return (
    <div>
      <h1>Data from Core Service</h1>
      <p><strong>AI Service Data:</strong> {aiServiceData}</p>
      <p><strong>PostgreSQL Data:</strong> {postgresData}</p>
    </div>
  );
}

export default App;

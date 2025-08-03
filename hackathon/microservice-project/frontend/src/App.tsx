import React from 'react';
import FeatureFlagExample from './components/FeatureFlagExample';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>Microservice Frontend</h1>
        <p>Feature Flag Integration Demo</p>
      </header>
      <main>
        <FeatureFlagExample />
      </main>
    </div>
  );
}

export default App;
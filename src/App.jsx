import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import AuthCard from './components/AuthCard';
import OnboardingModal from './components/OnboardingModal';

const App = () => {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<><AuthCard /><OnboardingModal /></>} />
        <Route path="/inbox" element={<div>Inbox Placeholder</div>} />
        <Route path="/" element={<><AuthCard /><OnboardingModal /></>} />
      </Routes>
    </Router>
  );
};

export default App; 
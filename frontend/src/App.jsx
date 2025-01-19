import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import MainPage from './views/MainPage';
import VerificationPage from './views/VerificationPage';
import Blogs from './views/Blogs';

function App() {
  const [user, setUser] = useState(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [loading, setLoading] = useState(true);

  // Fetch user info and determine login status
  const checkLoggedIn = async () => {
    try {
      const response = await fetch('http://localhost:9696/api/v1/user/getinfo', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });
      if (response.ok) {
        const data = await response.json();
        setUser(data);
        setIsLoggedIn(true);
      }
    } catch (error) {
      console.error('Error fetching user info:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkLoggedIn();
  }, []);

  if (loading) return <div>Loading...</div>;

  return (
    <Router>
      <Routes>

        <Route
          path="/"
          element={
            isLoggedIn
              ? user?.verified
                ? <Navigate to="/blogs" />
                : <Navigate to="/verification" />
              : <MainPage setIsLoggedIn={setIsLoggedIn} setUser={setUser} />
          }
        />

        <Route
          path="/verification"
          element={
            isLoggedIn
              ? user?.verified
                ? <Navigate to="/blogs" />
                : <VerificationPage user={user} setUser={setUser} />
              : <Navigate to="/" />
          }
        />

        <Route
          path="/blogs"
          element={
            isLoggedIn
              ? user?.verified
                ? <Blogs />
                : <Navigate to="/verification" />
              : <Navigate to="/" />
          }
        />
      </Routes>
    </Router>
  );
}

export default App;

import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import MainPage from './views/MainPage';
import VerificationPage from './views/VerificationPage';
import Blogs from './views/Blogs';

function App() {
  const [user, setUser] = useState(null);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [csrfToken, setCsrfToken] = useState("")
  const [loading, setLoading] = useState(true);

  const API_BASE_URL = import.meta.env.VITE_BACKEND_URL || '';

  // Fetch user info and determine login status
  const checkLoggedIn = async () => {
    try {
      const response = await fetch(API_BASE_URL + '/api/v1/user/getinfo', {
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
        setCsrfToken(response.headers.get("X-CSRF-Token"));
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
            isLoggedIn ? (
              user?.verified ? (
                <Navigate to="/blogs" />
              ) : (
                <Navigate to="/verification" />
              )
            ) : (
              <MainPage setIsLoggedIn={setIsLoggedIn} setUser={setUser} apiUrl={API_BASE_URL} />
            )
          }
        />

        <Route
          path="/verification"
          element={
            isLoggedIn ? (
              user?.verified ? (
                <Navigate to="/blogs" />
              ) : (
                <VerificationPage user={user} setUser={setUser} apiUrl={API_BASE_URL} csrfToken={csrfToken}/>
              )
            ) : (
              <Navigate to="/" />
            )
          }
        />

        <Route
          path="/blogs"
          element={
            isLoggedIn ? (
              user?.verified ? (
                <Blogs apiUrl={API_BASE_URL} csrfToken={csrfToken}/>
              ) : (
                <Navigate to="/verification" />
              )
            ) : (
              <Navigate to="/" />
            )
          }
        />
      </Routes>
    </Router>
  );
}

export default App;

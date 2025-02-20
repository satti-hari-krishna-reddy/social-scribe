import React, { useState } from 'react';
import { Box, Typography, TextField, Button, Modal } from '@mui/material';
import { useNavigate } from 'react-router-dom';

const SignUpModal = ({ open, handleClose, setIsLoggedIn, setUser, apiUrl }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSignUp = async() => {

    if (password !== confirmPassword) {
      alert('Passwords do not match!');
      return;
    }
    try {
      const response = await fetch(apiUrl + '/api/v1/user/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          username : email, 
          password: password 
        }),
        credentials: 'include',
      })
      if (response.status === 409) {
        setError('User already exists');
        return;
      }
      const data = await response.json();
      if (response.ok) {
     
      setUser(data);
      setIsLoggedIn(true);
      navigate('/verification');
    } }
      catch (error) {
        console.error('Error fetching user:', error);
        setUser(null);
        setError('Error occurred');
      }
  };
  return (
    <Modal
      open={open}
      onClose={handleClose}
      sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }} 
    >
      <Box 
        sx={{ 
          backgroundColor: '#2E2E2E', 
          borderRadius: '8px', 
          padding: '2rem', 
          width: { xs: '90%', sm: '400px' }, 
          boxShadow: 24 
        }}
      >
        <Typography variant="h6" gutterBottom sx={{ color: '#FF6B6B', textAlign: 'center' }}>
          Sign Up
        </Typography>

        <TextField
          label="Email"
          variant="outlined"
          fullWidth
          sx={{ marginBottom: '1rem' }}
          InputLabelProps={{
            style: { color: '#FFFFFF' }
          }}
          InputProps={{
            style: { color: '#FFFFFF' },
            placeholder: 'Enter your email',
          }}
          placeholder="Enter your email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />

        <TextField
          label="Password"
          type="password"
          variant="outlined"
          fullWidth
          sx={{ marginBottom: '1rem' }}
          InputLabelProps={{
            style: { color: '#FFFFFF' } 
          }}
          InputProps={{
            style: { color: '#FFFFFF' },
            placeholder: 'Enter your password',
          }}
          placeholder="Enter your password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />

        <TextField
          label="Confirm Password"
          type="password"
          variant="outlined"
          fullWidth
          sx={{ marginBottom: '1rem' }}
          InputLabelProps={{
            style: { color: '#FFFFFF' } 
          }}
          InputProps={{
            style: { color: '#FFFFFF' }, 
            placeholder: 'Re-enter your password',
          }}
          placeholder="Re-enter your password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
        />
        {error && <Typography color="red" variant="body2">{error}</Typography>}

        <Button 
          variant="contained" 
          onClick={handleSignUp} 
          sx={{ 
            backgroundColor: '#FF6B6B', 
            color: '#FFFFFF', 
            fontWeight: 'bold', 
            width: '100%' 
          }}
        >
          Sign Up
        </Button>

        <Typography 
          sx={{ 
            marginTop: '1rem', 
            textAlign: 'center', 
            color: '#FFFFFF' 
          }}
        >
          Already have an account? <span onClick={handleClose} style={{ color: '#FF6B6B', cursor: 'pointer' }}>Login</span>
        </Typography>
      </Box>
    </Modal>
  );
};

export default SignUpModal;

import React, { useState } from 'react';
import { Modal, Box, Typography, TextField, Button, IconButton } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import { useNavigate } from 'react-router-dom';

const Login = ({ open, handleClose, setUser, setIsLoggedIn }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async () => {
    try {
      const response = await fetch('http://localhost:9696/api/v1/user/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: email,
          password: password,
      }),
        credentials: 'include',
      });
      const data = await response.json();
      if (response.ok) {
        setUser(data);
        setIsLoggedIn(true);
        const verified = data?.verified;
        if (verified) {
          navigate('/blogs');
        } else {
          navigate('/verification');
        } 
  } else {
    setError(data.message);
  }

  }
  catch (error) {
    console.error('Error during login:', error);
    setError('Error during login');
  }}

  return (
    <Modal open={open} onClose={handleClose} sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      <Box 
        sx={{ 
          width: '100%', 
          maxWidth: 400, 
          padding: '2rem', 
          backgroundColor: '#2E2E2E', 
          borderRadius: '12px', 
          color: '#FFFFFF',
          boxShadow: 24,
          textAlign: 'center', 
          position: 'relative'
        }}
      >
        <IconButton 
          onClick={handleClose} 
          sx={{ position: 'absolute', top: 8, right: 8, color: '#FF6B6B' }}
        >
          <CloseIcon />
        </IconButton>
        
        <Typography variant="h5" sx={{ fontWeight: 'bold', marginBottom: '1rem', color: '#FF6B6B' }}>
          Login
        </Typography>

        <TextField
          label="Email"
          variant="outlined"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          fullWidth
          sx={{ 
            marginBottom: 2, 
            backgroundColor: '#3A3A3A', 
            borderRadius: '5px',
            input: { color: '#FFFFFF' }, 
            '& .MuiOutlinedInput-root': {
              '& fieldset': {
                borderColor: '#FF6B6B',
              },
              '&:hover fieldset': {
                borderColor: '#FF6B6B',
              },
              '&.Mui-focused fieldset': {
                borderColor: '#FF6B6B',
              },
            },
          }}
          InputLabelProps={{
            style: { color: '#FFFFFF' } 
          }}
        />
        <TextField
          label="Password"
          type="password"
          variant="outlined"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          fullWidth
          sx={{ 
            marginBottom: 2, 
            backgroundColor: '#3A3A3A', 
            borderRadius: '5px',
            input: { color: '#FFFFFF' },
            '& .MuiOutlinedInput-root': {
              '& fieldset': {
                borderColor: '#FF6B6B',
              },
              '&:hover fieldset': {
                borderColor: '#FF6B6B',
              },
              '&.Mui-focused fieldset': {
                borderColor: '#FF6B6B',
              },
            },
          }}
          InputLabelProps={{
            style: { color: '#FFFFFF' } 
          }}
        />
        {error && <Typography color="red" variant="body2">{error}</Typography>}
        
        <Button 
          variant="contained" 
          fullWidth
          sx={{ 
            marginTop: 2, 
            backgroundColor: '#FF6B6B', 
            color: 'white', 
            fontWeight: 'bold',
            '&:hover': {
              backgroundColor: '#FF4C4C',
            },
          }} 
          onClick={handleLogin}
        >
          Login
        </Button>

        <Typography 
          variant="body2" 
          sx={{ 
            marginTop: 2, 
            color: '#B8B8B8', 
            cursor: 'pointer',
            '&:hover': {
              textDecoration: 'underline',
            },
          }} 
          onClick={() => {}}
        >
          Forgot Password?
        </Typography>
      </Box>
    </Modal>
  );
};

export default Login;

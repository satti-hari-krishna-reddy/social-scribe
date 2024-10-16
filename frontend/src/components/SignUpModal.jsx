import React, { useState } from 'react';
import { Box, Typography, TextField, Button, Modal } from '@mui/material';

const SignUpModal = ({ open, handleClose }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const handleSignUp = () => {
    console.log("Sign Up Details", { email, password, confirmPassword });
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

import React, { useState } from 'react';
import { Modal, Box, Typography, TextField, Button, IconButton } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';

const Login = ({ open, handleClose, setUser, setIsLoggedIn }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [forgotMode, setForgotMode] = useState(false);
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
        if (data?.verified) {
          navigate('/blogs');
        } else {
          navigate('/verification');
        }
      } else if (response.status === 400) {
        setError(data.reason || 'Failed to login');
      } else if (response.status === 429) {
        setError(data.reason || 'Too Many Requests');
        toast.error('Too many requests. Wait for 1 minute before trying again');
      }
    } catch (err) {
      toast.error('Error during login');
    }
  };

  const handleForgotPassword = async () => {
    if (!email) {
      toast.error('Please enter your email');
      return;
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      toast.error('Please enter a valid email');
      return;
    }
    try {
      const response = await fetch('http://localhost:9696/api/v1/user/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });
      if (response.ok) {
        toast.success('OTP sent successfully');
        setForgotMode(true);
      }  else if (response.status === 429) {
        setError(data.reason || 'Too Many Requests');
        toast.error('Too many requests. Wait for 1 minute before trying again');
      } else {
        const data = await response.json();
        toast.error(data.message || 'Failed to send OTP');
      }
    } catch (err) {
      toast.error('Error sending OTP');
    }
  };

  const handleResetPassword = async () => {
    if (!otp || !newPassword || !confirmPassword) {
      toast.error('Please fill in all fields');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('Passwords do not match');
      return;
    }
    try {
      const response = await fetch('http://localhost:9696/api/v1/user/reset-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email,
          otp,
          password: newPassword,
        }),
      });
      if (response.status === 400) {
        toast.error('Invalid OTP');
        return;
      }
      if (response.status === 410) {
        toast.error('OTP expired');
        return;
      }
      if (response.status === 429) {
        setError(data.reason || 'Too Many Requests');
        toast.error('Too many requests. Wait for 1 minute before trying again');
        return;
      }
      if (response.ok) {
        toast.success('Password reset successfully');
        setForgotMode(false);
        setPassword(newPassword);
        handleClose();
      } else {
        const data = await response.json();
        toast.error(data.message || 'Failed to reset password');
      }
    } catch (err) {
      toast.error('Error resetting password');
    }
  };

  const handleResendOTP = async () => {
    if (!email) {
      toast.error('Email is required to resend OTP');
      return;
    }
    try {
      const response = await fetch('http://localhost:9696/api/v1/user/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });
      if (response.ok) {
        toast.success('OTP resent successfully');
      }  else if (response.status === 429) {
        setError(data.reason || 'Too Many Requests');
        toast.error('Too many requests. Wait for 1 minute before trying again');
      } else {
        const data = await response.json();
        toast.error(data.message || 'Failed to resend OTP');
      }
    } catch (err) {
      toast.error('Error resending OTP');
    }
  };

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
        <IconButton onClick={handleClose} sx={{ position: 'absolute', top: 8, right: 8, color: '#FF6B6B' }}>
          <CloseIcon />
        </IconButton>
        
        <Typography variant="h5" sx={{ fontWeight: 'bold', marginBottom: '1rem', color: '#FF6B6B' }}>
          {forgotMode ? 'Reset Password' : 'Login'}
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
              '& fieldset': { borderColor: '#FF6B6B' },
              '&:hover fieldset': { borderColor: '#FF6B6B' },
              '&.Mui-focused fieldset': { borderColor: '#FF6B6B' },
            },
          }}
          InputLabelProps={{ style: { color: '#FFFFFF' } }}
          disabled={forgotMode}
        />

        {forgotMode ? (
          <>
            <TextField
              label="OTP"
              variant="outlined"
              value={otp}
              onChange={(e) => setOtp(e.target.value)}
              fullWidth
              sx={{ 
                marginBottom: 2, 
                backgroundColor: '#3A3A3A', 
                borderRadius: '5px',
                input: { color: '#FFFFFF' },
                '& .MuiOutlinedInput-root': {
                  '& fieldset': { borderColor: '#FF6B6B' },
                  '&:hover fieldset': { borderColor: '#FF6B6B' },
                  '&.Mui-focused fieldset': { borderColor: '#FF6B6B' },
                },
              }}
              InputLabelProps={{ style: { color: '#FFFFFF' } }}
            />
            <TextField
              label="New Password"
              type="password"
              variant="outlined"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              fullWidth
              sx={{ 
                marginBottom: 2, 
                backgroundColor: '#3A3A3A', 
                borderRadius: '5px',
                input: { color: '#FFFFFF' },
                '& .MuiOutlinedInput-root': {
                  '& fieldset': { borderColor: '#FF6B6B' },
                  '&:hover fieldset': { borderColor: '#FF6B6B' },
                  '&.Mui-focused fieldset': { borderColor: '#FF6B6B' },
                },
              }}
              InputLabelProps={{ style: { color: '#FFFFFF' } }}
            />
            <TextField
              label="Confirm New Password"
              type="password"
              variant="outlined"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              fullWidth
              sx={{ 
                marginBottom: 2, 
                backgroundColor: '#3A3A3A', 
                borderRadius: '5px',
                input: { color: '#FFFFFF' },
                '& .MuiOutlinedInput-root': {
                  '& fieldset': { borderColor: '#FF6B6B' },
                  '&:hover fieldset': { borderColor: '#FF6B6B' },
                  '&.Mui-focused fieldset': { borderColor: '#FF6B6B' },
                },
              }}
              InputLabelProps={{ style: { color: '#FFFFFF' } }}
            />
          </>
        ) : (
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
                '& fieldset': { borderColor: '#FF6B6B' },
                '&:hover fieldset': { borderColor: '#FF6B6B' },
                '&.Mui-focused fieldset': { borderColor: '#FF6B6B' },
              },
            }}
            InputLabelProps={{ style: { color: '#FFFFFF' } }}
          />
        )}
        
        {error && <Typography color="red" variant="body2">{error}</Typography>}
        
        <Button 
          variant="contained" 
          fullWidth
          sx={{ 
            marginTop: 2, 
            backgroundColor: '#FF6B6B', 
            color: 'white', 
            fontWeight: 'bold',
            '&:hover': { backgroundColor: '#FF4C4C' },
          }} 
          onClick={forgotMode ? handleResetPassword : handleLogin}
        >
          {forgotMode ? 'Reset Now' : 'Login'}
        </Button>

        {forgotMode ? (
          <>
            <Button 
              variant="text" 
              fullWidth
              sx={{ marginTop: 1, color: '#FF6B6B' }}
              onClick={handleResendOTP}
            >
              Resend OTP
            </Button>
            <Typography 
              variant="body2" 
              sx={{ 
                marginTop: 2, 
                color: '#B8B8B8', 
                cursor: 'pointer',
                '&:hover': { textDecoration: 'underline' },
              }} 
              onClick={() => setForgotMode(false)}
            >
              Back to Login
            </Typography>
          </>
        ) : (
          <Typography 
            variant="body2" 
            sx={{ 
              marginTop: 2, 
              color: '#B8B8B8', 
              cursor: 'pointer',
              '&:hover': { textDecoration: 'underline' },
            }} 
            onClick={handleForgotPassword}
          >
            Forgot Password?
          </Typography>
        )}
      </Box>
    </Modal>
  );
};

export default Login;

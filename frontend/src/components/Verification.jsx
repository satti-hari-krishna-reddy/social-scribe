import React, { useState } from 'react';
import { Box, Typography, Button } from '@mui/material';

const EmailVerification = ({ resendVerificationEmail }) => {
  const [resendMessage, setResendMessage] = useState('');

  const handleResend = () => {
    resendVerificationEmail();
    setResendMessage('A verification email has been sent to your email address.');
  };

  return (
    <Box
      sx={{
        width: '100%',
        height: '100vh',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: '#2E2E2E',
        color: '#FFFFFF',
      }}
    >
      <Box
        sx={{
          textAlign: 'center',
          padding: '2rem',
          backgroundColor: '#3A3A3A',
          borderRadius: '10px',
        }}
      >
        <Typography
          variant="h5"
          sx={{ color: '#FF6B6B', fontWeight: 'bold', marginBottom: '1rem' }}
        >
          Verify Your Email
        </Typography>
        <Typography sx={{ marginBottom: '1rem', color: '#FFFFFF' }}>
          A verification link has been sent to your email. Please verify your account before
          continuing.
        </Typography>
        <Button
          variant="contained"
          onClick={handleResend}
          sx={{
            backgroundColor: '#FF6B6B',
            color: '#FFFFFF',
            fontWeight: 'bold',
            marginBottom: '1rem',
          }}
        >
          Resend Verification Email
        </Button>
        {resendMessage && (
          <Typography sx={{ color: '#FF6B6B', marginTop: '1rem' }}>{resendMessage}</Typography>
        )}
      </Box>
    </Box>
  );
};

export default EmailVerification;

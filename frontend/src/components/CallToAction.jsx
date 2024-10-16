import React from 'react';
import { Box, Typography, Button } from '@mui/material';

const CallToAction = () => {
  return (
    <Box sx={{ backgroundColor: '#1A1A1', padding: '3rem', textAlign: 'center' }}>
      <Typography variant="h4" gutterBottom sx={{ color: '#FFC107' }}>
        Start Automating Your Blog Sharing
      </Typography>
      <Button variant="contained" sx={{ backgroundColor: '#FF6B6B', color: '#FFFFFF', marginRight: '1rem' }}>
        Sign Up
      </Button>
      <Button variant="outlined" sx={{ borderColor: '#FF6B6B', color: '#FFFFFF' }}>
        Log In
      </Button>
    </Box>
  );
};

export default CallToAction;

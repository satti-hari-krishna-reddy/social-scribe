import React from 'react';
import { Box, Typography } from '@mui/material';

const CallToAction = () => {
  return (
    <Box sx={{ backgroundColor: '#1A1A1', padding: '3rem', textAlign: 'center' }}>
      <Typography variant="h4" gutterBottom sx={{ color: '#FFC107' }}>
        Start Automating Your Blog Sharing
      </Typography>
    </Box>
  );
};

export default CallToAction;

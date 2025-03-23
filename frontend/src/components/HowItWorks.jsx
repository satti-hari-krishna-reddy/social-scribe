import React from 'react';
import { Box, Typography } from '@mui/material';

const HowItWorks = () => {
  return (
    <Box sx={{ padding: '3rem 2rem', backgroundColor: '#1A1A1A', color: '#FFFFFF' }}>
      <Typography
        variant="h4"
        gutterBottom
        sx={{ textAlign: 'center', fontWeight: 'bold', color: '#FFC107' }}
      >
        How It Works
      </Typography>
      <Typography variant="body1" sx={{ textAlign: 'center', maxWidth: '800px', margin: '0 auto' }}>
        Connect your Hashnode, X, LinkedIn, and ChatGPT accounts. Automate blog summarization and
        scheduling with a few clicks.
      </Typography>
    </Box>
  );
};

export default HowItWorks;

import React from 'react';
import { Box, Typography } from '@mui/material';

const About = () => {
  return (
    <Box sx={{ padding: '3rem 2rem', backgroundColor: '#1A1A1A', color: '#FFFFFF' }}>
      <Typography variant="h4" gutterBottom sx={{ textAlign: 'center', fontWeight: 'bold', color: '#FF6B6B' }}>
        About Social Scribe
      </Typography>
      <Typography 
        variant="body1" 
        sx={{ textAlign: 'center', maxWidth: '800px', margin: '0 auto', lineHeight: 1.6, fontWeight: 'bold'}}
      >
        Social Scribe is a personal project that originated from an idea I had while participating in the Hashnode API Hackathon. 
        Although it didn’t end up winning, the concept stayed with me, and I decided to bring it to life. 
        The goal was simple: automate sharing Hashnode blogs across X (Twitter) and LinkedIn, while adding features like scheduling 
        and AI-powered summaries.
      </Typography>
      <Typography 
        variant="body1" 
        sx={{ textAlign: 'center', maxWidth: '800px', margin: '1rem auto', lineHeight: 1.6, fontWeight: 'bold' }}
      >
        What started as a hackathon idea is now something I continue to work on and improve, even if it's just for personal use 
        or to demonstrate what can be built with APIs, automation, and AI.
      </Typography>
    </Box>
  );
};

export default About;

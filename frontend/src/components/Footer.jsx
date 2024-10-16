import React from 'react';
import { Box, Typography } from '@mui/material';

const Footer = () => {
  const currentYear = new Date().getFullYear(); 

  return (
    <Box sx={{ padding: '1rem', backgroundColor: '#2E2E2E', color: '#FFFFFF', textAlign: 'center' }}>
      <Typography variant="body2">
        Social Scribe © {currentYear} | Automate sharing your Hashnode blogs across X and LinkedIn.
      </Typography>
    </Box>
  );
};

export default Footer;

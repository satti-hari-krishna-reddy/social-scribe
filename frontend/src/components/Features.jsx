import React from 'react';
import { Box, Typography, Grid } from '@mui/material';

const Features = () => {
  return (
    <Box sx={{ padding: '3rem 2rem', backgroundColor: '#2E2E2E', color: '#FFFFFF' }}>
      <Typography
        variant="h4"
        gutterBottom
        sx={{ textAlign: 'center', fontWeight: 'bold', color: '#FF6B6B' }}
      >
        Key Features
      </Typography>
      <Grid container spacing={4}>
        <Grid item xs={12} sm={4}>
          <Box
            sx={{ padding: '2rem', backgroundColor: '#393939', borderRadius: '12px', boxShadow: 1 }}
          >
            <Typography variant="h5" gutterBottom sx={{ fontWeight: 'bold', color: '#FFC107' }}>
              Real-time Sharing
            </Typography>
            <Typography variant="body1">
              Instantly share your Hashnode blogs across X and LinkedIn the moment they are
              published.
            </Typography>
          </Box>
        </Grid>

        <Grid item xs={12} sm={4}>
          <Box
            sx={{ padding: '2rem', backgroundColor: '#393939', borderRadius: '12px', boxShadow: 1 }}
          >
            <Typography variant="h5" gutterBottom sx={{ fontWeight: 'bold', color: '#FFC107' }}>
              Scheduling Posts
            </Typography>
            <Typography variant="body1">
              Use our scheduling feature to plan when your blog posts will be shared, ensuring
              optimal timing.
            </Typography>
          </Box>
        </Grid>

        <Grid item xs={12} sm={4}>
          <Box
            sx={{ padding: '2rem', backgroundColor: '#393939', borderRadius: '12px', boxShadow: 1 }}
          >
            <Typography variant="h5" gutterBottom sx={{ fontWeight: 'bold', color: '#FFC107' }}>
              AI-Powered Summaries
            </Typography>
            <Typography variant="body1">
              Use ChatGPT to create short, engaging summaries before sharing to socials.
            </Typography>
          </Box>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Features;

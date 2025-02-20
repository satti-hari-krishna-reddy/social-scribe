import React from 'react';
import { Grid, Container, Typography, Box } from '@mui/material';
import BlogCard from './BlogCard';
import NotificationBell from './Notifications';

const BlogGrid = ({ blogs, loading, apiUrl }) => {
  // If there's exactly one blog, render it with a fixed width container.
  if (loading) {
    return (
      <>
        <NotificationBell />
        <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              minHeight: '50vh',
            }}
          >
            <Typography variant="h6" color="white">
              Loading blogs...
            </Typography>
          </Box>
        </Container>
      </>
    );
  }

  if (blogs.length === 0) {
    return (
      <>
        <NotificationBell />
        <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              minHeight: '50vh',
            }}
          >
            <Typography variant="h6" color="white">
              Dude, there are no blogs to show!
            </Typography>
          </Box>
        </Container>
      </>
    );
  }

  if (blogs.length === 1) {
    return (
      <>
        <NotificationBell />
        <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              minHeight: '50vh',
            }}
          >
            <Box sx={{ width: 320 }}>
              <BlogCard blog={blogs[0]} apiUrl={apiUrl} />
            </Box>
          </Box>
        </Container>
      </>
    );
  }

  // For more than one blog, use the Grid layout as usual.
  return (
    <>
      <NotificationBell />
      <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
        <Grid container spacing={4} justifyContent="center">
          {blogs.map((blog, index) => (
            <Grid item xs={12} sm={6} md={4} key={index}>
              <BlogCard blog={blog} apiUrl={apiUrl} />
            </Grid>
          ))}
        </Grid>
      </Container>
    </>
  );
};

export default BlogGrid;

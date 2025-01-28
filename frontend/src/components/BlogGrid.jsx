import React from 'react';
import { Grid, Container, Typography } from '@mui/material';
import BlogCard from './BlogCard';
import NotificationBell from './Notifications';

const BlogGrid = ({ blogs, loading }) => {
  return (
    <>
      <NotificationBell />
      <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
        {loading ? (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              minHeight: '50vh',
            }}
          >
            <Typography variant="h6" color="white">
              Loading blogs...
            </Typography>
          </div>
        ) : blogs.length === 0 ? (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              minHeight: '50vh',
            }}
          >
            <Typography variant="h6" color="white">
              Dude, there are no blogs to show!
            </Typography>
          </div>
        ) : (
          <Grid container spacing={4}>
            {blogs.map((blog, index) => (
              <Grid item xs={12} sm={6} md={4} key={index}>
                <BlogCard blog={blog} />
              </Grid>
            ))}
          </Grid>
        )}
      </Container>
    </>
  );
};

export default BlogGrid;

import React from 'react';
import { Grid, Container } from '@mui/material';
import BlogCard from './BlogCard';
import NotificationBell from './Notifications';
import BlogSectionTabs from './BlogSection';

const BlogGrid = ({ blogs }) => {
  return (
    <>
<NotificationBell />
<BlogSectionTabs />
    <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
      <Grid container spacing={4}>
        {blogs.map((blog, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <BlogCard blog={blog} />
          </Grid>
        ))}
      </Grid>
    </Container>
    </>
  );
};

export default BlogGrid;

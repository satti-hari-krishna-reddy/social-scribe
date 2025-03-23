import React, { useState } from 'react';
import { Grid, Container, Typography, Box, Button, Modal } from '@mui/material';
import BlogCard from './BlogCard';
import NotificationBell from './Notifications';
import { toast } from 'react-toastify';

const modalStyle = {
  position: 'absolute',
  top: '50%',
  left: '50%',
  transform: 'translate(-50%, -50%)',
  bgcolor: '#2c2c2c',
  color: 'white',
  borderRadius: 2,
  boxShadow: 24,
  p: 4,
  width: 300,
};

const BlogGrid = ({ blogs, loading, apiUrl }) => {
  const [openLogoutModal, setOpenLogoutModal] = useState(false);

  const handleLogout = async () => {
    try {
      const response = await fetch(`${apiUrl}/api/v1/user/logout`, {
        method: 'POST',
        credentials: 'include',
      });
      if (response.ok) {
        toast.success('Log out successful');
        setTimeout(() => window.location.reload(), 1000);
      } else {
        toast.error('Error logging out');
      }
    } catch (error) {
      toast.error('Error logging out');
    }
    setOpenLogoutModal(false);
  };

  const renderContent = () => {
    if (loading) {
      return (
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
      );
    }

    if (blogs.length === 0) {
      return (
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
      );
    }

    if (blogs.length === 1) {
      return (
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
      );
    }

    return (
      <Grid container spacing={4} justifyContent="center">
        {blogs.map((blog, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <BlogCard blog={blog} apiUrl={apiUrl} />
          </Grid>
        ))}
      </Grid>
    );
  };

  return (
    <>
      <Container maxWidth="lg" sx={{ marginTop: '30px' }}>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            mb: 2,
          }}
        >
          {/* <Button
            variant="contained"
            color="error"
            onClick={() => setOpenLogoutModal(true)}
            sx={{ mr: 0.5 }} // Reduced margin to bring it closer to the notification bell
          >
            Log Out
          </Button> */}
          <NotificationBell />
        </Box>
        {renderContent()}
      </Container>
      <Modal
        open={openLogoutModal}
        onClose={() => setOpenLogoutModal(false)}
        aria-labelledby="logout-modal-title"
        aria-describedby="logout-modal-description"
      >
        <Box sx={modalStyle}>
          <Typography id="logout-modal-title" variant="h6" component="h2">
            Confirm Logout
          </Typography>
          <Typography id="logout-modal-description" sx={{ mt: 2 }}>
            Are you sure you want to log out?
          </Typography>
          <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
            <Button variant="contained" onClick={() => setOpenLogoutModal(false)}>
              Cancel
            </Button>
            <Button variant="contained" color="error" onClick={handleLogout}>
              Yes, Log Out
            </Button>
          </Box>
        </Box>
      </Modal>
    </>
  );
};

export default BlogGrid;

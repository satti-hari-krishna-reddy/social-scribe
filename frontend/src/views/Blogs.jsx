import React from 'react';
import { useSearchParams } from 'react-router-dom';
import BlogGrid from '../components/BlogGrid';
import BlogSectionTabs from '../components/BlogSection';
import { CircularProgress, Box, Button, Modal, Typography, TextField } from '@mui/material'; // Added Modal & Typography here
import { toast } from 'react-toastify'; // Only add if it's missing and used

const Blogs = ({ apiUrl }) => {
  const [blogs, setBlogs] = React.useState([]);
  const [loading, setLoading] = React.useState(false);
  const [searchParams, setSearchParams] = useSearchParams();
  const [openLogoutModal, setOpenLogoutModal] = React.useState(false);
  const [openDeleteAccountModal, setOpenDeleteAccountModal] = React.useState(false);
  const [password, setPassword] = React.useState('');

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

  const handleAccountDelete = async () => {
    if (!password) {
      toast.error('Please enter your password');
      return;
    }

    try {
      const response = await fetch(`${apiUrl}/api/v1/user/delete-account`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ password: password }),
      });
      if (response.ok) {
        toast.success('Account Deleted successfully');
        setTimeout(() => window.location.reload(), 500);
      } else if (response.status == 401) {
        toast.error('password is incorrect');
      }
    } catch (error) {
      toast.error('Error Deleting account');
      setOpenDeleteAccountModal(false);
    }
  };

  const getBlogs = async (tab) => {
    setLoading(true);
    try {
      const response = await fetch(apiUrl + `/api/v1/user/blogs?category=${tab}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (!response.ok) {
        const error = await response.json();
        console.error('Failed to fetch blogs:', error.message || 'Unknown error');
        setBlogs([]);
        setLoading(false);
        return;
      }

      const { blogs } = await response.json();
      setBlogs(blogs || []);
    } catch (err) {
      console.error('An error occurred while fetching blogs:', err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleTabChange = (tab) => {
    setSearchParams({ tab });
    getBlogs(tab);
  };

  React.useEffect(() => {
    const tab = searchParams.get('tab') || 'all';
    getBlogs(tab);
  }, [searchParams]);

  // Dark-themed modal styling
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

  return (
    <div style={{ backgroundColor: '#2E2E2E', minHeight: '100vh', padding: '20px' }}>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexDirection: 'row',
        }}
      >
        <BlogSectionTabs
          activeTab={searchParams.get('tab') || 'all'}
          onTabChange={handleTabChange}
        />
        <Button
          variant="contained"
          color="error"
          onClick={() => setOpenLogoutModal(true)}
          sx={{
            ml: 0.1,
            mr: 8,
            backgroundColor: '#FF6B6B',
            '&:hover': { backgroundColor: '#e05555' },
          }} // Added a bit of left margin
        >
          Log Out
        </Button>
        <Button
          variant="contained"
          color="error"
          onClick={() => setOpenDeleteAccountModal(true)}
          sx={{ ml: 0.1, mr: 8 }} // Added a bit of left margin
        >
          Delete Account
        </Button>
      </Box>

      {loading ? (
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
          <CircularProgress />
        </Box>
      ) : blogs.length > 0 ? (
        <BlogGrid blogs={blogs} apiUrl={apiUrl} />
      ) : (
        <Box
          display="flex"
          justifyContent="center"
          alignItems="center"
          minHeight="50vh"
          color="white"
        >
          <p>Dude, there are no blogs here!</p>
        </Box>
      )}

      {/* Logout Confirmation Modal */}
      <Modal open={openLogoutModal} onClose={() => setOpenLogoutModal(false)}>
        <Box sx={modalStyle}>
          <Typography variant="h6">Confirm Logout</Typography>
          <Typography sx={{ mt: 2 }}>Are you sure you want to log out?</Typography>
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

      {/* Delete Account Confirmation Modal */}
      <Modal
        open={openDeleteAccountModal}
        onClose={() => setOpenDeleteAccountModal(false)}
        aria-labelledby="delete-account-modal-title"
        aria-describedby="delete-account-modal-description"
      >
        <Box
          sx={{
            ...modalStyle,
            p: 4,
            bgcolor: '#1E1E1E',
            borderRadius: '8px',
            boxShadow: 24,
            maxWidth: 500,
            mx: 'auto',
          }}
        >
          <Typography
            id="delete-account-modal-title"
            variant="h6"
            component="h2"
            sx={{ fontWeight: 'bold', color: '#FFFFFF' }}
          >
            Confirm Account Deletion
          </Typography>
          <Typography id="delete-account-modal-description" sx={{ mt: 2, color: '#CCCCCC' }}>
            Are you sure you want to delete your account? This action is irreversible, and all your
            data will be permanently erased.
          </Typography>
          <TextField
            label="Enter Password"
            type="password"
            variant="outlined"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            fullWidth
            sx={{
              mt: 3,
              backgroundColor: '#3A3A3A',
              borderRadius: '5px',
              input: { color: '#FFFFFF' },
              '& .MuiOutlinedInput-root': {
                '& fieldset': { borderColor: '#FF6B6B' },
                '&:hover fieldset': { borderColor: '#FF6B6B' },
                '&.Mui-focused fieldset': { borderColor: '#FF6B6B' },
              },
            }}
            InputLabelProps={{ style: { color: '#FFFFFF' } }}
          />
          <Box sx={{ mt: 3, display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
            <Button
              variant="contained"
              onClick={() => setOpenDeleteAccountModal(false)}
              sx={{
                backgroundColor: '#6C757D',
                '&:hover': { backgroundColor: '#5A6268' },
              }}
            >
              Cancel
            </Button>
            <Button variant="contained" color="error" onClick={handleAccountDelete}>
              Yes, Delete
            </Button>
          </Box>
        </Box>
      </Modal>
    </div>
  );
};

export default Blogs;

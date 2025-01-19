import React, { useState } from 'react';
import { Box, Typography, Button } from '@mui/material';
import Login from './LoginModal'; 
import SignUpModal from './SignUpModal'; 

const HeroSection = ({setIsLoggedIn, setUser}) => {
  const [openLogin, setOpenLogin] = useState(false); 
  const [openSignUp, setOpenSignUp] = useState(false); 

  return (
    <Box 
      sx={{ 
        backgroundColor: '#2E2E2E', 
        padding: { xs: '2rem', sm: '4rem' }, 
        color: '#FFFFFF', 
        textAlign: 'center' 
      }}
    >
      <Typography 
        variant="h2" 
        gutterBottom 
        sx={{ fontWeight: 'bold', color: '#FF6B6B', fontSize: { xs: '1.8rem', md: '2.5rem' } }} 
      >
        Social Scribe: Automate Your Hashnode Blogs Sharing
      </Typography>
      <Typography 
        variant="h6" 
        gutterBottom 
        sx={{ 
          maxWidth: { xs: '90%', sm: '700px' }, 
          margin: '0 auto', 
          color: '#FFFFFF' 
        }}
      >
        Connect Hashnode, LinkedIn, X, and ChatGPT to seamlessly summarize, share, and schedule your blogs across all platforms.
      </Typography>
      
      <Box sx={{ display: 'flex', justifyContent: 'center', gap: '1rem', marginTop: '2rem' }}>
        {/* Login button to open the modal */}
        <Button 
          variant="contained" 
          onClick={() => setOpenLogin(true)} 
          sx={{ 
            backgroundColor: '#FF6B6B', 
            color: '#FFFFFF', 
            fontWeight: 'bold', 
            padding: '0.75rem 2rem' 
          }}
        >
          Login
        </Button>

        {}
        <Button 
          variant="contained" 
          onClick={() => setOpenSignUp(true)} 
          sx={{ 
            backgroundColor: '#FF6B6B', 
            color: '#FFFFFF', 
            fontWeight: 'bold', 
            padding: '0.75rem 2rem' 
          }}
        >
          Sign Up
        </Button>
      </Box>
      
      <Login open={openLogin} handleClose={() => setOpenLogin(false)} setUser={setUser} setIsLoggedIn={setIsLoggedIn} />
      <SignUpModal open={openSignUp} handleClose={() => setOpenSignUp(false)} setIsLoggedIn={setIsLoggedIn} setUser={setUser} />
    </Box>
  );
};

export default HeroSection;

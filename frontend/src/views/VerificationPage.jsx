import React, { useState, useEffect } from "react";
import {
  Box,
  Typography,
  TextField,
  Button,

} from "@mui/material";
import TwitterIcon from "@mui/icons-material/Twitter";
import LinkedInIcon from "@mui/icons-material/LinkedIn";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";

const VerificationPage = () => {
  const [twitterConnected, setTwitterConnected] = useState(null);
  const [linkedinConnected, setLinkedinConnected] = useState(null);
  const [hashnodeVerified, setHashnodeVerified] = useState(false);
  const [emailVerified, setEmailVerified] = useState("");
  const [email, setEmail] = useState("hari66.hks@gmail.com");
  const [hashnodeApiKey, setHashnodeApiKey] = useState("");
  const [otp, setOtp] = useState("");
  const [otpStatus, setOtpStatus] = useState(null);
  const [disabled, setDisabled] = useState(true);

  const handleTwitterConnect = () => {
    // Simulate API call
    setTwitterConnected(true);
  };

  const handleLinkedInConnect = () => {
    // Simulate API call
    setLinkedinConnected(true);
  };

  const handleHashnodeVerify = () => {
    // Simulate API call
    if (hashnodeApiKey) {
      setHashnodeVerified(true);
    }
  };

  const handleOtpVerify = () => {
    // Simulate OTP verification
    if (otp === "123456") {
        setEmailVerified(true);
      setOtpStatus("success");
    } else {
      setOtpStatus("failed");
    }
  };

  const handleResendOtp = () => {
    // Simulate OTP resend
    setOtp("");
    setOtpStatus(null);
  };

  useEffect(() => { 
    if (emailVerified && hashnodeVerified && (linkedinConnected || twitterConnected )) {
        setDisabled(false);
        }
    }, [emailVerified, hashnodeVerified, linkedinConnected, twitterConnected]);




  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "#2E2E2E",
        color: "#FFFFFF",
        padding: "2rem",
      }}
    >
      <Box
        sx={{
          width: "100%",
          maxWidth: "500px",
          padding: "2rem",
          backgroundColor: "#393939",
          borderRadius: "12px",
          boxShadow: 3,
        }}
      >
        <Typography
          variant="h5"
          gutterBottom
          sx={{ fontWeight: "bold", color: "#FFC107", marginBottom: "1rem" }}
        >
          Account Verification
        </Typography>

        {/* Twitter Connect */}
        <Box sx={{ marginBottom: "1.5rem" }}>
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Box display="flex" alignItems="center" gap="0.5rem">
              <TwitterIcon sx={{ color: "#1DA1F2" }} />
              <Typography>Connect X (Twitter)</Typography>
              {twitterConnected ? (<CheckCircleIcon color="success" />) : <></>}
            </Box>
            {twitterConnected ? (
              <Button
                variant="contained"
                color="error"
              >
                Disconnect
              </Button>
            ) : (
              <Button variant="contained" onClick={handleTwitterConnect}>
                Connect
              </Button>
            )}
          </Box>
        </Box>

        {/* LinkedIn Connect */}
        <Box sx={{ marginBottom: "1.5rem" }}>
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Box display="flex" alignItems="center" gap="0.5rem">
              <LinkedInIcon sx={{ color: "#0077B5" }} />
              <Typography>Connect LinkedIn</Typography>
              {linkedinConnected ? (<CheckCircleIcon color="success" />) : <></>}
            </Box>
            {linkedinConnected ? (
              <Button
                variant="contained"
                color="error"
              >
                Disconnect
              </Button>
            ) : (
              <Button variant="contained" onClick={handleLinkedInConnect}>
                Connect
              </Button>
            )}
          </Box>
        </Box>

        {/* Hashnode API Key */}
        <Box sx={{ marginBottom: "1.5rem" }}>
        <Box display="flex" alignItems="center" justifyContent="space-between">
        <Box display="flex" alignItems="center" gap="0.5rem">
          <Typography>Hashnode API Key</Typography>
          {hashnodeVerified ? (
              <CheckCircleIcon color="success" />
            ) : <></>}
            </Box>
          <Box display="flex" alignItems="center" gap="1rem">
          {hashnodeVerified ? ( <Button variant="contained"  color="error" onClick={handleHashnodeVerify}>
                Reset Key
              </Button> ) : (
                <>
            <TextField
              variant="outlined"
              size="small"
              fullWidth
              disabled={hashnodeVerified}
              value={hashnodeApiKey}
              onChange={(e) => setHashnodeApiKey(e.target.value)}
            />
              <Button variant="contained" onClick={handleHashnodeVerify}>
                verify
              </Button>
                </>
            )}
          </Box>
        </Box>
        </Box>

        {/* Verify Email */}
        <Box> 
            {emailVerified ? (
                <Box display="flex" alignItems="center" gap="1rem" marginBottom="0.5rem">
                 <Typography>{email}</Typography>
                <CheckCircleIcon color="success" /> 
                </Box>
            ) : <>
          <Typography>Verify your Email</Typography>
          <Box display="flex" alignItems="center" gap="1rem" marginBottom="0.5rem">
            <TextField
              variant="outlined"
              size="small"
              fullWidth
              value={otp}
              onChange={(e) => setOtp(e.target.value)}
              placeholder="Enter OTP"
            />
            <Button variant="contained" onClick={handleOtpVerify}>
              Verify
            </Button>
          </Box>
          <Button
            variant="text"
            color="secondary"
            onClick={handleResendOtp}
          >
            Resend OTP
          </Button>
          {otpStatus === "success" && (
            <Typography color="success.main" marginTop="0.5rem">
              OTP Verified Successfully
            </Typography>
          )}
          {otpStatus === "failed" && (
            <Typography color="error" marginTop="0.5rem">
              OTP Verification Failed
            </Typography>
          )}
</>}
        </Box>
        <Button sx={{ color: 'black', backgroundColor: 'white', marginTop : '20px', marginLeft : '450px' }} disabled={disabled}  onClick={handleHashnodeVerify}>
                next
              </Button>
      </Box>

    </Box>
  );
};

export default VerificationPage;

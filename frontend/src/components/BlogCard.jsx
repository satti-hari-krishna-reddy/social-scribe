import React, { useState } from 'react';
import { useLocation } from 'react-router-dom';
import {
  Card,
  CardMedia,
  CardContent,
  Typography,
  CardActionArea,
  Button,
  Box,
  IconButton,
  Popover,
  Modal,
  TextField,
  Checkbox,
  FormControlLabel,
} from '@mui/material';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import CloseIcon from '@mui/icons-material/Close';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { toast } from "react-toastify"; 
import dayjs from 'dayjs';

const BlogCard = ({ blog }) => {
  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);
  const tab = queryParams.get('tab');
  
  const [anchorEl, setAnchorEl] = useState(null);
  const [openSchedule, setOpenSchedule] = useState(false);
  const [selectedDate, setSelectedDate] = useState(dayjs());
  const [shareOptions, setShareOptions] = useState({
    linkedin: false,
    x: false,
  });

  const handleShareClick = (event) => setAnchorEl(event.currentTarget);
  const handleCloseShare = () => setAnchorEl(null);
  const handleOpenSchedule = () => setOpenSchedule(true);
  const handleCloseSchedule = () => setOpenSchedule(false);
  const handleDateChange = (newDate) => setSelectedDate(newDate);

  const handleShareOptionChange = (event) => {
    setShareOptions({ ...shareOptions, [event.target.name]: event.target.checked });
  };

  const handleConfirmShare = async () => {
    const platforms = [];
    if (shareOptions.linkedin) platforms.push("linkedin");
    if (shareOptions.x) platforms.push("twitter");
  
    if (platforms.length === 0) {
      toast.warning("Please select at least one platform to share!");
      return;
    }
  
    try {
      const response = await fetch("http://localhost:9696/api/v1/blogs/user/share", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ id: blog.id, platforms }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to share the blog");
      }

      toast.success("Blog shared successfully!");
    } catch (error) {
      console.error("Error sharing the blog:", error.message);
      toast.error(error.message || "Failed to share the blog. Please try again!");
    } finally {
      handleCloseShare();
    }
  };

  const handleCancelSchedule = () => {
    // Implement your cancel schedule logic here
    console.log('Cancel schedule for:', blog.id);
    toast.info("Schedule canceled");
  };

  const open = Boolean(anchorEl);
  const id = open ? 'share-popover' : undefined;

  return (
    <Card sx={{ backgroundColor: '#1A1A1A', color: 'white', borderRadius: '10px', boxShadow: 3, position: 'relative' }}>
      <CardActionArea href={blog.url} target="_blank" rel="noopener noreferrer">
        <CardMedia component="img" height="180" image={blog.coverImage.url} alt={blog.title} />
        <CardContent>
          <Typography gutterBottom variant="h6" component="div" sx={{ fontSize: '18px', fontWeight: 'bold', color: '#FFFFFF' }}>
            {blog.title}
          </Typography>
          <Typography variant="body2" sx={{ fontSize: '14px', color: '#B8B8B8' }}>
            {blog.author.name}
          </Typography>
          <div style={{ display: 'flex', alignItems: 'center', marginTop: '10px' }}>
            <AccessTimeIcon sx={{ fontSize: '16px', color: '#909090', marginRight: '5px' }} />
            <Typography variant="body2" sx={{ fontSize: '14px', color: '#909090' }}>
              {blog.readTimeInMinutes} min read
            </Typography>
          </div>
        </CardContent>
      </CardActionArea>

      <Box sx={{ 
        display: 'flex', 
        justifyContent: 'flex-end', 
        padding: '10px', 
        gap: '8px', 
        position: 'absolute', 
        bottom: 0, 
        right: 0 
      }}>
        {tab === 'shared' ? (
          <Typography variant="body2" sx={{ 
            color: '#909090',
            padding: '6px 12px',
            fontSize: '14px',
            alignSelf: 'center'
          }}>
            Shared on {dayjs(blog.shared_time).format('MMM D, YYYY [at] HH:mm')}
          </Typography>
        ) : tab === 'scheduled' ? (
          <Button
            variant="outlined"
            sx={{ 
              borderColor: '#ff4444', 
              color: '#ff4444', 
              '&:hover': { borderColor: '#cc0000' },
              fontSize: '12px', 
              textTransform: 'none', 
              minWidth: '60px', 
              padding: '5px' 
            }}
            onClick={handleCancelSchedule}
          >
            Cancel
          </Button>
        ) : (
          <>
            <Button
              variant="contained"
              sx={{ 
                backgroundColor: 'white', 
                color: 'black', 
                fontSize: '12px', 
                textTransform: 'none', 
                minWidth: '60px', 
                padding: '5px',
                '&:hover': { backgroundColor: '#f0f0f0' }
              }}
              onClick={handleShareClick}
            >
              Share
            </Button>

            <Popover
              id={id}
              open={open}
              anchorEl={anchorEl}
              onClose={handleCloseShare}
              anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
            >
              <Box sx={{ 
                padding: '10px', 
                display: 'flex', 
                flexDirection: 'column', 
                backgroundColor: '#2E2E2E' 
              }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body1" sx={{ fontWeight: 'bold', color: 'white' }}>Share to:</Typography>
                  <IconButton size="small" onClick={handleCloseShare} sx={{ color: 'white' }}>
                    <CloseIcon />
                  </IconButton>
                </Box>
                <FormControlLabel
                  control={<Checkbox sx={{ color: 'white' }} checked={shareOptions.linkedin} onChange={handleShareOptionChange} name="linkedin" />}
                  label={<Typography sx={{ color: 'white' }}>LinkedIn</Typography>}
                />
                <FormControlLabel
                  control={<Checkbox sx={{ color: 'white' }} checked={shareOptions.x} onChange={handleShareOptionChange} name="x" />}
                  label={<Typography sx={{ color: 'white' }}>X</Typography>}
                />
                <Button
                  variant="contained"
                  color="primary"
                  sx={{ 
                    marginTop: '10px',
                    backgroundColor: '#1976d2',
                    '&:hover': { backgroundColor: '#1565c0' }
                  }}
                  onClick={handleConfirmShare}
                >
                  Share Now
                </Button>
              </Box>
            </Popover>

            <Button
              variant="outlined"
              sx={{ 
                borderColor: 'white', 
                color: 'white', 
                fontSize: '12px', 
                textTransform: 'none', 
                minWidth: '60px', 
                padding: '5px',
                '&:hover': { borderColor: '#e0e0e0' }
              }}
              onClick={handleOpenSchedule}
            >
              Schedule
            </Button>

            <Modal open={openSchedule} onClose={handleCloseSchedule}>
              <Box sx={{ 
                width: 400, 
                padding: '20px', 
                margin: 'auto', 
                marginTop: '10%', 
                backgroundColor: '#2E2E2E', 
                borderRadius: '10px', 
                color: 'white' 
              }}>
                <Typography variant="h6" gutterBottom>
                  Schedule Post
                </Typography>
                <LocalizationProvider dateAdapter={AdapterDayjs}>
                  <DateTimePicker
                    label="Select Date & Time"
                    value={selectedDate}
                    onChange={handleDateChange}
                    disablePast
                    maxDate={dayjs().add(7, 'day')}
                    format="YYYY-MM-DD HH:mm"
                    renderInput={(props) => (
                      <TextField 
                        {...props} 
                        fullWidth 
                        sx={{ 
                          backgroundColor: '#424242', 
                          borderRadius: '5px',
                          '& .MuiInputBase-input': { color: 'white' }
                        }} 
                      />
                    )}
                  />
                </LocalizationProvider>
                <Button
                  variant="contained"
                  color="primary"
                  sx={{ 
                    marginTop: '20px', 
                    width: '100%',
                    backgroundColor: '#1976d2',
                    '&:hover': { backgroundColor: '#1565c0' }
                  }}
                  onClick={handleCloseSchedule}
                >
                  Confirm Schedule
                </Button>
              </Box>
            </Modal>
          </>
        )}
      </Box>
    </Card>
  );
};

export default BlogCard;
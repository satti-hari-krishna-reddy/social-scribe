import React, { useState } from 'react';
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
  const handleCloseSchedule = () => {setOpenSchedule(false); 
    console.log(selectedDate)
  }
  const handleDateChange = (newDate) => {setSelectedDate(newDate);
    console.log(newDate)
  }

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
  
    const requestBody = {
      id: blog.id,
      platforms,
    };
  
    try {
      const response = await fetch("http://localhost:9696/api/v1/blogs/user/share", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(requestBody),
      });
  
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || "Failed to share the blog");
      }
  
      toast.success("Blog shared successfully!");
    } catch (error) {
      console.error("Error sharing the blog:", error.message);
      toast.error("Failed to share the blog. Please try again!");
    } finally {
      handleCloseShare(); 
    }
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

      <Box sx={{ display: 'flex', justifyContent: 'flex-end', padding: '10px', gap: '8px', position: 'absolute', bottom: 0, right: 0 }}>
 
        <Button
          variant="contained"
          sx={{ backgroundColor: 'white', color: 'black', fontSize: '12px', textTransform: 'none', minWidth: '60px', padding: '5px' }}
          onClick={handleShareClick}
        >
          Share
        </Button>

   
        <Popover
          id={id}
          open={open}
          anchorEl={anchorEl}
          onClose={handleCloseShare}
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'left',
          }}
        >
          <Box sx={{ padding: '10px', display: 'flex', flexDirection: 'column', backgroundColor: 'black' }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between'}}>
              <Typography variant="body1" sx={{ fontWeight: 'bold', color: 'white' }}>Share to:</Typography>
              <IconButton size="small" onClick={handleCloseShare}>
                <CloseIcon />
              </IconButton>
            </Box>
            <FormControlLabel
              control={<Checkbox sx={{ color: 'white' }} checked={shareOptions.linkedin} onChange={handleShareOptionChange} name="linkedin" />}
              label={
                <Typography sx={{ color: 'white' }}>LinkedIn</Typography>
              }
            />
            <FormControlLabel
              control={<Checkbox sx={{ color: 'white' }} checked={shareOptions.x} onChange={handleShareOptionChange} name="x" />}
              label={
                <Typography sx={{ color: 'white' }}>X</Typography>
              }
            />
            <Button
              variant="contained"
              color="secondary"
              sx={{ marginTop: '10px' }}
              onClick={handleConfirmShare}
            >
              OK
            </Button>
          </Box>
        </Popover>

      
        <Button
          variant="outlined"
          sx={{ borderColor: 'white', color: 'white', fontSize: '12px', textTransform: 'none', minWidth: '60px', padding: '5px' }}
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
                maxDate={dayjs().add(7, 'day')} // Max 7 days from now
                format="YYYY-MM-DD HH:mm" // Set to 24-hour format
                renderInput={(props) => <TextField {...props} fullWidth sx={{ backgroundColor: 'white', borderRadius: '5px' }} />}
              />
            </LocalizationProvider>
            <Button
              variant="contained"
              color="secondary"
              sx={{ marginTop: '20px', width: '100%' }}
              onClick={handleCloseSchedule}
            >
              Confirm Schedule
            </Button>
          </Box>
        </Modal>
      </Box>
    </Card>
  );
};

export default BlogCard;

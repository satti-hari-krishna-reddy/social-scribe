import React, { useState } from 'react';
import {
  Badge,
  IconButton,
  Modal,
  Typography,
  Box,
  Button,
  List,
  ListItem,
  Divider,
} from '@mui/material';
import NotificationsIcon from '@mui/icons-material/Notifications';

const NotificationBell = () => {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState(['You have no new notifications!']);

  const handleOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);
  const clearNotifications = () => setNotifications([]);

  return (
    <div style={{ position: 'absolute', top: 20, right: 20 }}>
      {' '}
      {}
      <IconButton onClick={handleOpen} color="inherit">
        <Badge badgeContent={notifications.length} color="secondary">
          <NotificationsIcon />
        </Badge>
      </IconButton>
      <Modal open={open} onClose={handleClose}>
        <Box
          sx={{
            width: 400,
            padding: 2,
            margin: 'auto',
            marginTop: '15%',
            backgroundColor: '#2E2E2E',
            borderRadius: '10px',
          }}
        >
          <Typography variant="h6" gutterBottom>
            Notifications
          </Typography>

          {notifications.length === 0 ? (
            <Typography>No notifications</Typography>
          ) : (
            <div>
              <List>
                {notifications.map((notif, index) => (
                  <React.Fragment key={index}>
                    <ListItem style={{ color: '#fff' }}>{notif}</ListItem>
                    {index < notifications.length - 1 && (
                      <Divider style={{ backgroundColor: '#555' }} />
                    )}{' '}
                    {}
                  </React.Fragment>
                ))}
              </List>
              <Button
                onClick={clearNotifications}
                variant="contained"
                color="secondary"
                fullWidth
                style={{ marginTop: '10px' }}
              >
                Clear Notifications
              </Button>
            </div>
          )}
        </Box>
      </Modal>
    </div>
  );
};

export default NotificationBell;

import React, { useState } from 'react';
import { Tabs, Tab } from '@mui/material';

const BlogSectionTabs = () => {
  const [value, setValue] = useState(0);

  const handleChange = (event, newValue) => {
    setValue(newValue);
  };

  return (
    <div style={{ marginTop: '20px' }}> 
      <Tabs
        value={value}
        onChange={handleChange}
        indicatorColor="secondary"
        textColor="white"
        centered
      >
        <Tab label="All Blogs" />
        <Tab label="Scheduled Blogs" />
        <Tab label="Shared Blogs" />
      </Tabs>
    </div>
  );
};

export default BlogSectionTabs;

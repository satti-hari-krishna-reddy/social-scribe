import React from "react";
import { Tabs, Tab } from "@mui/material";

const BlogSectionTabs = ({ activeTab, onTabChange }) => {
  const handleChange = (event, newValue) => {
    onTabChange(newValue);
  };

  return (
    <div style={{ marginTop: "20px" }}>
      <Tabs
        value={activeTab}
        onChange={handleChange}
        indicatorColor="secondary"
        textColor="white"
        centered
      >
        <Tab label="All Blogs" value="all" />
        <Tab label="Scheduled Blogs" value="scheduled" />
        <Tab label="Shared Blogs" value="shared" />
      </Tabs>
    </div>
  );
};

export default BlogSectionTabs;

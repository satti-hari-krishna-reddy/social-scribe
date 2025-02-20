import React from "react";
import { useSearchParams } from "react-router-dom";
import BlogGrid from "../components/BlogGrid";
import BlogSectionTabs from "../components/BlogSection";
import { CircularProgress, Box } from "@mui/material";

const Blogs = ({apiUrl}) => {
  const [blogs, setBlogs] = React.useState([]);
  const [loading, setLoading] = React.useState(false);
  const [searchParams, setSearchParams] = useSearchParams();

  const getBlogs = async (tab) => {
    setLoading(true);
    try {
      const response = await fetch(apiUrl + `/api/v1/user/blogs?category=${tab}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
      });

      if (!response.ok) {
        const error = await response.json();
        console.error("Failed to fetch blogs:", error.message || "Unknown error");
        setBlogs([]);
        setLoading(false);
        return;
      }

      const { blogs } = await response.json();
      setBlogs(blogs || []);
    } catch (err) {
      console.error("An error occurred while fetching blogs:", err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleTabChange = (tab) => {
    setSearchParams({ tab }); 
    getBlogs(tab);
  };

  React.useEffect(() => {
    const tab = searchParams.get("tab") || "all"; 
    getBlogs(tab);
  }, [searchParams]);

  return (
    <div style={{ backgroundColor: "#2E2E2E", minHeight: "100vh", padding: "20px" }}>
      <BlogSectionTabs
        activeTab={searchParams.get("tab") || "all"} 
        onTabChange={handleTabChange}
      />
      {loading ? (
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
          <CircularProgress />
        </Box>
      ) : blogs.length > 0 ? (
        <BlogGrid blogs={blogs} apiUrl={apiUrl}/>
      ) : (
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh" color="white">
          <p>Dude, there are no blogs here!</p>
        </Box>
      )}
    </div>
  );
};

export default Blogs;

import React from 'react';
import BlogGrid from '../components/BlogGrid';

// const blogs = [
//   {
//     node: {
//       title: "Building on Bitcoin Layers With the Hiro Platform",
//       url: "https://blog.developerdao.com/building-on-bitcoin-layers-with-the-hiro-platform",
//       id: "6621743cbe424d2d968e6c0f",
//       coverImage: {
//         url: "https://cdn.hashnode.com/res/hashnode/image/upload/v1712227564078/85f5c303-d030-4057-b87f-63a3b28cb2ed.jpeg"
//       },
//       author: {
//         name: "Ϗ"
//       },
//       readTimeInMinutes: 10
//     }
//   },
//   {
//     node: {
//       title: "Comparing Solidity With Clarity",
//       url: "https://blog.developerdao.com/comparing-solidity-with-clarity",
//       id: "661653ad8d11bcf90baed36e",
//       coverImage: {
//         url: "https://cdn.hashnode.com/res/hashnode/image/upload/v1712218211483/e641392a-484a-4fc6-9d98-826a286dbd39.jpeg"
//       },
//       author: {
//         name: "Ϗ"
//       },
//       readTimeInMinutes: 9
//     }
//   },
//   {
//     node: {
//       title: "Creating a Token-Gated Web Page With Clarity",
//       url: "https://blog.developerdao.com/creating-a-token-gated-web-page-with-clarity",
//       id: "660468b71b0b20cc9a2b4e9e",
//       coverImage: {
//         url: "https://cdn.hashnode.com/res/hashnode/image/upload/v1711563528738/b3c252fc-f857-4428-8df5-59c72c4a20d1.jpeg"
//       },
//       author: {
//         name: "Osikhena Oshomah"
//       },
//       readTimeInMinutes: 26
//     }
//   },
//   {
//     node: {
//       title: "Farcaster Frames Explained",
//       url: "https://blog.developerdao.com/farcaster-frames-explained",
//       id: "65cf724e6ad6a686aaccd1a3",
//       coverImage: {
//         url: "https://cdn.hashnode.com/res/hashnode/image/upload/v1708093430682/45c9d98e-94dd-4e72-a51b-7853498834c1.jpeg"
//       },
//       author: {
//         name: "Ϗ"
//       },
//       readTimeInMinutes: 9
//     }
//   },
//   {
//     node: {
//       title: "What is Quadratic Funding, and How Does it Work?",
//       url: "https://blog.developerdao.com/what-is-quadratic-funding",
//       id: "65ccc4ccdfe21d142ce011bf",
//       coverImage: {
//         url: "https://cdn.hashnode.com/res/hashnode/image/upload/v1707918392643/d8eb2433-1cb3-43f3-b8ae-4c64cf926d21.jpeg"
//       },
//       author: {
//         name: "Victor Fawole"
//       },
//       readTimeInMinutes: 8
//     }
//   },
//   {
//     node: {
//       title: "Crafting Beautiful Light Art with Figma and Midjourney",
//       url: "https://blog.developerdao.com/crafting-beautiful-light-art-with-figma-and-midjourney",
//       id: "65ae2f5dd03d0ce559281cd3",
//       coverImage: {
//         url: "https://cdn.hashnode.com/res/hashnode/image/upload/v1705913981276/9ed107a6-5d9a-4a16-9128-af4025696943.jpeg"
//       },
//       author: {
//         name: "Erik Knobl"
//       },
//       readTimeInMinutes: 4
//     }
//   }
// ];


const Blogs = () => {
  const [blogs, setBlogs] = React.useState([]);

  React.useEffect(() => {
    getBlogs();
  }, []);

  const getBlogs = async () => {
    try {
      const response = await fetch('http://localhost:9696/api/v1/user/blogs', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });
  
      if (!response.ok) {
        const error = await response.json();
        console.error('Failed to fetch blogs:', error.message || 'Unknown error');
        return;
      }
  
      const { blogs } = await response.json();
      setBlogs(blogs || []);
    } catch (err) {
      console.error('An error occurred while fetching blogs:', err.message);
    }
  };
  
  return (
    <div style={{ backgroundColor: '#2E2E2E', minHeight: '100vh', padding: '20px' }}>
      <BlogGrid blogs={blogs} />
    </div>
  );
}

export default Blogs;

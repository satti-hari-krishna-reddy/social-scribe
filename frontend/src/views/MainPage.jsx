import React from 'react';
import HeroSection from '../components/HeroSection';
import HowItWorks from '../components/HowItWorks';
import Features from '../components/Features';
import CallToAction from '../components/CallToAction';
import Footer from '../components/Footer';
import About from '../components/About';

function MainPage({ setIsLoggedIn, setUser, apiUrl }) {
    
  return (
    <div>
      <HeroSection setIsLoggedIn={setIsLoggedIn} setUser={setUser} apiUrl={apiUrl} />
      <HowItWorks />
      <CallToAction />
      <Features />
      <About />
      <Footer />
    </div>
  );
}

export default MainPage;
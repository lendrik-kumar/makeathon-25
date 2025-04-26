import React from 'react';
import SplineModel from '../../components/splineModel';

const Landing = () => {
  return (
    <div className="relative w-full min-h-screen overflow-hidden bg-black text-white">
      {/* Background Spline Model */}
      <div className="absolute inset-0 z-0">
        <SplineModel />
      </div>

      {/* Foreground Content */}
      <div className="relative z-10 p-8 flex flex-col items-center">
        {/* Heading */}
        <h1 className="text-5xl font-bold text-center mb-10 leading-tight">
          Get the <span className="text-green-400">Application</span> you<br />Want for Growthhj
        </h1>

        {/* Search Bar */}
        <div className="w-full max-w-xl bg-white bg-opacity-10 backdrop-blur-md rounded-full p-3 flex items-center mb-16">
          <input
            type="text"
            placeholder="Search API, Apps & Plugin"
            className="bg-transparent outline-none text-white w-full px-4 placeholder-white"
          />
          <span className="px-4">⌘F</span>
        </div>

        {/* Marketplace Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 w-full max-w-6xl">
          {/* Card 1 */}
          <div className="p-6 rounded-2xl bg-white bg-opacity-10 backdrop-blur-md flex flex-col gap-4">
            <h2 className="text-xl font-semibold">File Manager</h2>
            <p>Total Users: 5.2k</p>
            <p>Downloads: 9,04,012+</p>
            <button className="bg-green-500 text-black font-semibold px-4 py-2 rounded-full w-max mt-auto">
              Download App
            </button>
          </div>

          {/* Card 2 */}
          <div className="p-6 rounded-2xl bg-white bg-opacity-10 backdrop-blur-md flex flex-col gap-4">
            <h2 className="text-xl font-semibold">Analytics Data</h2>
            <p>Total Users: 9.2k</p>
            <p>Downloads: 1,00,000+</p>
            <button className="bg-yellow-400 text-black font-semibold px-4 py-2 rounded-full w-max mt-auto">
              Download App
            </button>
          </div>

          {/* Card 3 */}
          <div className="p-6 rounded-2xl bg-white bg-opacity-10 backdrop-blur-md flex flex-col gap-4">
            <h2 className="text-xl font-semibold">Wallet Feature</h2>
            <p>Total Users: 4.8k</p>
            <p>Downloads: 70,800+</p>
            <button className="bg-blue-400 text-black font-semibold px-4 py-2 rounded-full w-max mt-auto">
              Download App
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Landing;
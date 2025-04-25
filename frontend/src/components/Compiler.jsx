import React from 'react'
import { Outlet } from 'react-router-dom'
import Navbar from './ui/Navbar'
import Footer from './ui/Footer'

function Compiler() {
  return (
    <>
    <Navbar />
    <Outlet />
    <Footer />
    </>
  )
}

export default Compiler
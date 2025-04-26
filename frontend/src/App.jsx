import SplineModel from "./components/splineModel"
import { BrowserRouter as Router , Routes , Route } from "react-router-dom"
import Compiler from './components/Compiler'
import Landing from "./pages/landing"


function App() {

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Compiler/>}>
          <Route path="Home" element={<Landing/>}></Route>
        </Route>
      </Routes>
    </Router>
  )
}

export default App

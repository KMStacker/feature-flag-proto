import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [flag, setFlag] = useState(false)

  useEffect(() => {
    const syncWithBackend = () => {
      fetch("http://localhost:8080/api/flags")
        .then(res => res.json())
        .then(data => {
          console.log("Feature Flag status is here:", data)
          setFlag(data["feature-flag-1"])
        })
        .catch(err => {
          console.error("Error getting data:", err)
        })
    }

    syncWithBackend()
    const interval = setInterval(syncWithBackend, 5000) // 5s

    return () => clearInterval(interval)
  }, [])

  /*
  const toggleFlag = () => {
    setFlag(prevValue => !prevValue)
  }
  */

  const toggleFlag = () => {
    const newValue = !flag

    fetch("http://localhost:8080/api/flags", {
      method: 'POST',
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ state: newValue }), // { state: true/false }
    })
    .then(res => {
      if (res.ok) {
        console.log("Synced with backend with succes!")
        setFlag(newValue)
      } else {
        console.error("Backend gave me some hard time for sure")
      }
    })
    .catch(err => console.error("Error:", err))
  }

  return (
    <div style={{ padding: '50px', textAlign: 'center', border: '5px solid #333' }}>
      <h1>This is for the flag setters!</h1>
      
      <div style={{ margin: '50px', transform: 'scale(1.5)' }}>
        <label style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '20px' }}>
          <span>Test Feature Number Uno</span>
          <input 
            type='checkbox' 
            checked={flag} 
            onChange={toggleFlag} 
            style={{ width: '25px', height: '25px', cursor: 'pointer' }}
          />
          <span style={{ color: flag ? 'green' : 'red' }}>
            {flag ? "ON" : "OFF"}
          </span>
        </label>
      </div>
    </div>
  )
}

export default App
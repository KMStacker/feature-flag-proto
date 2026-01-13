import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [flag, setFlag] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const checkFlag = () => {
      fetch('http://localhost:8080/api/flags')
        .then(res => res.json())
        .then(data => {
          console.log("Data is here:", data)
          setFlag(data["feature-flag-1"])
          setLoading(false)
        })
        .catch(err => {
          console.error("Error getting data:", err)
          setLoading(false)
        })
    }

    checkFlag()
    const interval = setInterval(checkFlag, 5000) // 5s

    return () => clearInterval(interval)
  }, [])

  

  if (loading) return <p>Loadinggg...</p>

  return (
    <div style={{ padding: '50px', textAlign: 'center' }}>
      <h1>My first feature flag test!</h1>
     
      {flag ? (
        <div style={{ backgroundColor: '#6eb17dff', padding: '20px', borderRadius: '10px' }}>
          <h2>✨ New fine feature is on FIRE! ✨</h2>
          <p>You see this because backend said "true".</p>
        </div>
      ) : (
        <div style={{ backgroundColor: '#a56066ff', padding: '20px', borderRadius: '10px' }}>
          <h2>Nada features here at all.. blaah...</h2>
          <p>You see this in case flag is "false".</p>
        </div>
      )}
    </div>
  )
}

export default App
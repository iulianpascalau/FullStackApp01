import { useState, useEffect } from 'react'
import './App.css'
import Login from './Login'

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
  const [role, setRole] = useState<string | null>(localStorage.getItem('role'))

  const [count, setCount] = useState<number | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string>('')
  const [showChangePassword, setShowChangePassword] = useState(false)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [passwordMessage, setPasswordMessage] = useState('')

  useEffect(() => {
    if (token) {
      fetchCounter()
    } else {
      setLoading(false)
    }
  }, [token])

  const handleLogin = (newToken: string, newRole: string) => {
    localStorage.setItem('token', newToken)
    localStorage.setItem('role', newRole)
    setToken(newToken)
    setRole(newRole)
    setLoading(true)
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    setToken(null)
    setRole(null)
    setCount(null)
    setError('')
    setLoading(false)
    setShowChangePassword(false)
    setOldPassword('')
    setNewPassword('')
    setPasswordMessage('')
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setPasswordMessage('')
    try {
      const response = await fetch('http://localhost:8080/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          old_password: oldPassword,
          new_password: newPassword
        })
      })

      if (response.ok) {
        setPasswordMessage('Password changed successfully!')
        setOldPassword('')
        setNewPassword('')
        setTimeout(() => setShowChangePassword(false), 1500)
      } else {
        const text = await response.text()
        setPasswordMessage(`Error: ${text}`)
      }
    } catch (err) {
      console.error(err)
      setPasswordMessage(`Failed to connect to server: ${err instanceof Error ? err.message : String(err)}`)
    }
  }

  const fetchCounter = async () => {
    try {
      const response = await fetch('http://localhost:8080/counter', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (response.status === 401) {
        handleLogout()
        return
      }
      if (!response.ok) {
        throw new Error('Failed to fetch counter')
      }
      const data = await response.json()
      setCount(data.value)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  const handleIncrement = async () => {
    try {
      // Optimistic update
      setCount((prev) => (prev !== null ? prev + 1 : 1))

      const response = await fetch('http://localhost:8080/counter', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) {
        throw new Error('Failed to increment')
      }
      const data = await response.json()
      setCount(data.value)
    } catch (err) {
      setError('Failed to update counter')
      // Revert on error (fetching real state)
      fetchCounter()
    }
  }

  const handleReset = async () => {
    try {
      setCount(0)
      const response = await fetch('http://localhost:8080/counter', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) {
        throw new Error('Failed to reset')
      }
      const data = await response.json()
      setCount(data.value)
    } catch (err) {
      setError('Failed to reset counter')
      fetchCounter()
    }
  }

  if (!token) {
    return (
      <div className="app-container">
        <header>
          <h1>Antigravity Demo</h1>
          <p className="subtitle">Secure Counter Access</p>
        </header>
        <Login onLogin={handleLogin} />
      </div>
    )
  }

  return (
    <div className="app-container">
      <div className="card">
        <header>
          <h1>LevelDB Counter</h1>
          <p className="subtitle">Logged in as {role}</p>
        </header>

        <div className="content">
          {showChangePassword ? (
            <div className="change-password-form">
              <h3>Change Password</h3>
              <form onSubmit={handleChangePassword}>
                <div className="form-group">
                  <input
                    type="password"
                    placeholder="Old Password"
                    value={oldPassword}
                    onChange={(e) => setOldPassword(e.target.value)}
                    required
                  />
                </div>
                <div className="form-group">
                  <input
                    type="password"
                    placeholder="New Password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    required
                  />
                </div>
                {passwordMessage && <p className="message">{passwordMessage}</p>}
                <div className="button-group">
                  <button type="submit" className="increment-btn">Update</button>
                  <button type="button" className="reset-btn" onClick={() => setShowChangePassword(false)}>Cancel</button>
                </div>
              </form>
            </div>
          ) : (
            <>
              {loading && <div className="loading-spinner"></div>}

              {error && (
                <div className="error-message">
                  <span>⚠️</span>
                  <p>{error}</p>
                </div>
              )}

              {!loading && !error && count !== null && (
                <div className="success-message">
                  <span className="count-display">{count}</span>
                  <div className="button-group">
                    {role === 'admin' && (
                      <button className="reset-btn" onClick={handleReset}>
                        Reset
                      </button>
                    )}
                    <button className="increment-btn" onClick={handleIncrement}>
                      Increment +
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>

        <div className="footer-actions">
          {!showChangePassword && (
            <button className="change-pass-btn" onClick={() => setShowChangePassword(true)}>
              Change Password
            </button>
          )}
          <button className="logout-btn" onClick={handleLogout}>Authorization: Logout</button>
        </div>

        <footer>
          <p>Powered by Go, LevelDB & React 19</p>
        </footer>
      </div>
    </div>
  )
}

export default App

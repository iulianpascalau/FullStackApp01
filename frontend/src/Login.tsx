import { useState } from 'react'
import './Login.css'

interface OrderProps {
    onLogin: (token: string, role: string) => void
}

export default function Login({ onLogin }: OrderProps) {
    const [isRegistering, setIsRegistering] = useState(false)
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError('')
        const endpoint = isRegistering ? '/register' : '/login'

        try {
            const response = await fetch(`${endpoint}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            })

            if (!response.ok) {
                const text = await response.text()
                throw new Error(text || 'Action failed')
            }

            if (isRegistering) {
                // After register, auto-login or simple switch
                setIsRegistering(false)
                alert('Registration successful! Please login.')
            } else {
                const data = await response.json()
                onLogin(data.token, data.role)
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'An error occurred')
        }
    }

    return (
        <div className="login-container">
            <div className="card login-card">
                <h2>{isRegistering ? 'Create Account' : 'Welcome Back'}</h2>
                <form onSubmit={handleSubmit}>
                    <div className="form-group">
                        <label>Username</label>
                        <input
                            type="text"
                            value={username}
                            onChange={e => setUsername(e.target.value)}
                            required
                        />
                    </div>
                    <div className="form-group">
                        <label>Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={e => setPassword(e.target.value)}
                            required
                        />
                    </div>
                    {error && <p className="error-text">{error}</p>}
                    <button type="submit" className="primary-btn">
                        {isRegistering ? 'Sign Up' : 'Login'}
                    </button>
                </form>
                <p className="switch-mode">
                    {isRegistering ? 'Already have an account?' : "Don't have an account?"}
                    <button
                        className="link-btn"
                        onClick={() => setIsRegistering(!isRegistering)}
                    >
                        {isRegistering ? 'Login' : 'Register'}
                    </button>
                </p>
            </div>
        </div>
    )
}

// frontend/src/pages/Login.jsx
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './Login.module.css';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const handleSubmit = (e) => {
    e.preventDefault();
    navigate('/report');
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1 className={styles.header}>Login to ASTU Lost & Found</h1>
        
        <form onSubmit={handleSubmit} className={styles.form}>
          <input
            type="email"
            placeholder="Email (student@astu.edu)"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={styles.input}
            required
          />
          
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className={styles.input}
            required
          />
          
          <button type="submit" className={styles.button}>
            Login
          </button>
        </form>
        
        <div className={styles.info}>
          <p>Student account: student@astu.edu / any password</p>
          <p>Admin account: admin@astu.edu / any password</p>
        </div>
      </div>
    </div>
  );
}
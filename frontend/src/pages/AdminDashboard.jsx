import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';
import styles from './Home.module.css';

const AdminDashboard = () => {
    const { user } = useAuth();
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        fetchAdminItems();
    }, []);

    const fetchAdminItems = async () => {
        setLoading(true);
        try {
            // Admin uses the same endpoint but sees more or has different control
            // Here we just fetch all items for the platform management
            const response = await api.get('/items');
            setItems(response.data);
            setError('');
        } catch (err) {
            console.error('Failed to fetch items', err);
            setError('Failed to load system items.');
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id) => {
        if (!window.confirm('Are you sure you want to delete this item?')) return;
        try {
            await api.delete(`/admin/items/${id}`);
            setItems(items.filter(item => item.id !== id));
        } catch (err) {
            alert('Failed to delete item: ' + (err.response?.data?.error || err.message));
        }
    };

    if (loading && items.length === 0) return <div className={styles.loading}>Loading Admin Console...</div>;

    return (
        <div className={styles.container}>
            <header className={styles.header}>
                <h1>Admin Command Center</h1>
                <p>Global view of all lost and found reports</p>
            </header>

            {error && <div className={styles.error}>{error}</div>}

            <section className={styles.stats}>
                <div className={styles.statCard}>
                    <h3>Total Items</h3>
                    <p>{items.length}</p>
                </div>
            </section>

            <div className={styles.tableContainer}>
                <table className={styles.table}>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Title</th>
                            <th>Type</th>
                            <th>Status</th>
                            <th>Reporter ID</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {items.map(item => (
                            <tr key={item.id}>
                                <td>{item.id}</td>
                                <td>{item.title}</td>
                                <td>
                                    <span className={item.type === 'lost' ? styles.tagLost : styles.tagFound}>
                                        {item.type.toUpperCase()}
                                    </span>
                                </td>
                                <td>{item.status}</td>
                                <td>{item.reporter_id}</td>
                                <td>
                                    <button 
                                        onClick={() => handleDelete(item.id)} 
                                        className={styles.deleteBtn}
                                    >
                                        Delete
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default AdminDashboard;

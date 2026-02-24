// frontend/src/pages/Report.jsx
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

export default function Report() {
  const [activeTab, setActiveTab] = useState('lost');
  const [formData, setFormData] = useState({
    title: '',
    location: '',
    description: '',
    image: null,
    category: 'other' // ADDED: category field
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(''); // ADDED: error state
  const navigate = useNavigate();
  const { user } = useAuth();

  // FIX 1: Check if user exists before rendering
  if (!user) {
    navigate('/login');
    return null;
  }

  const handleInputChange = (e) => {
    const { name, value, files } = e.target;
    if (name === 'image') {
      // FIX 2: Better file handling with validation
      const file = files[0];
      if (file) {
        // Validate file size (max 5MB)
        if (file.size > 5 * 1024 * 1024) {
          setError('Image size must be less than 5MB');
          return;
        }
        // Validate file type
        if (!file.type.startsWith('image/')) {
          setError('Please upload an image file');
          return;
        }
        setFormData(prev => ({ ...prev, image: file }));
        setError(''); // Clear error
      }
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');
    
    try {
      // FIX 3: Validate required fields
      if (!formData.title.trim()) {
        throw new Error('Item name is required');
      }
      if (!formData.location.trim()) {
        throw new Error('Location is required');
      }

      // Create FormData for file upload
      const submitData = new FormData();
      submitData.append('type', activeTab);
      submitData.append('title', formData.title.trim());
      submitData.append('location', formData.location.trim());
      submitData.append('description', formData.description.trim() || 'No description provided');
      submitData.append('category', formData.category);
      submitData.append('user_id', user.id.toString()); // FIX 4: Ensure string
      submitData.append('date_reported', new Date().toISOString());
      
      if (formData.image) {
        submitData.append('image', formData.image);
      }
      
      // FIX 5: Remove Content-Type header - let browser set it with boundary
      await api.post('/api/items', submitData, {
        headers: { 
          // Don't set Content-Type here - browser will set it with boundary
        }
      });
      
      alert(`✅ Successfully reported ${activeTab} item!`);
      navigate('/dashboard'); // FIX 6: Navigate to dashboard, not '/dashboard'
      
    } catch (error) {
      console.error('Submission error:', error);
      setError(error.message || 'Failed to submit report. Please try again.');
    } finally {
      setSubmitting(false);
    }
  };

  // FIX 7: Reset form function
  const handleReset = () => {
    setFormData({
      title: '',
      location: '',
      description: '',
      image: null,
      category: 'other'
    });
    setError('');
  };

  return (
    <div style={styles.container}>
      <header style={styles.header}>
        <h1>Report {activeTab === 'lost' ? 'Lost' : 'Found'} Item</h1>
        <div>
          <button onClick={handleReset} style={styles.resetButton}>
            Reset
          </button>
          <button onClick={() => navigate('/')} style={styles.homeButton}>
            ← Back
          </button>
        </div>
      </header>

      {/* FIX 8: Show error message */}
      {error && (
        <div style={styles.error}>
          {error}
        </div>
      )}

      <div style={styles.tabs}>
        <button 
          onClick={() => {
            setActiveTab('lost');
            setError(''); // Clear error on tab switch
          }}
          style={activeTab === 'lost' ? styles.activeTab : styles.tab}
        >
          🟢 Report Lost Item
        </button>
        <button 
          onClick={() => {
            setActiveTab('found');
            setError('');
          }}
          style={activeTab === 'found' ? styles.activeTab : styles.tab}
        >
          🔵 Report Found Item
        </button>
      </div>

      <form onSubmit={handleSubmit} style={styles.form}>
        {/* FIX 9: Add category field */}
        <div style={styles.formGroup}>
          <label>Category *</label>
          <select 
            name="category"
            value={formData.category}
            onChange={handleInputChange}
            style={styles.select}
            required
          >
            <option value="electronics">Electronics</option>
            <option value="id_card">ID Card</option>
            <option value="books">Books</option>
            <option value="clothing">Clothing</option>
            <option value="accessories">Accessories</option>
            <option value="other">Other</option>
          </select>
        </div>

        <div style={styles.formGroup}>
          <label>Item Name *</label>
          <input 
            name="title"
            value={formData.title}
            onChange={handleInputChange}
            placeholder="e.g., HP Calculator, Student ID Card"
            required 
            style={styles.input} 
          />
        </div>
        
        <div style={styles.formGroup}>
          <label>Location *</label>
          <input 
            name="location"
            value={formData.location}
            onChange={handleInputChange}
            placeholder="e.g., Library, Room 201, Cafeteria"
            required 
            style={styles.input} 
          />
        </div>
        
        <div style={styles.formGroup}>
          <label>Description</label>
          <textarea 
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            placeholder="Color, brand, distinguishing features..."
            style={{...styles.input, height: '100px'}} 
          />
        </div>
        
        <div style={styles.formGroup}>
          <label>Photo (optional, max 5MB)</label>
          <input 
            type="file" 
            name="image"
            accept="image/jpeg,image/png,image/gif" 
            onChange={handleInputChange}
            style={styles.fileInput} 
          />
          {/* FIX 10: Show selected filename */}
          {formData.image && (
            <small style={styles.fileName}>
              Selected: {formData.image.name}
            </small>
          )}
        </div>
        
        <div style={styles.buttonGroup}>
          <button 
            type="submit" 
            style={styles.submitButton}
            disabled={submitting}
          >
            {submitting ? '⏳ Submitting...' : `✅ Submit ${activeTab === 'lost' ? 'Lost' : 'Found'} Report`}
          </button>
        </div>
      </form>

      {/* FIX 11: Add helpful tips */}
      <div style={styles.tips}>
        <h4>📝 Tips:</h4>
        <ul style={styles.tipsList}>
          <li>Be specific about the location where item was lost/found</li>
          <li>Include unique features in description</li>
          <li>Clear photos help identify items faster</li>
        </ul>
      </div>
    </div>
  );
}


const styles = {
  container: {
    maxWidth: '800px',
    margin: '0 auto',
    padding: '20px'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '20px 0',
    marginBottom: '20px',
    borderBottom: '2px solid #1a365d'
  },
  homeButton: {
    background: '#f1f5f9',
    border: 'none',
    padding: '8px 15px',
    borderRadius: '4px',
    cursor: 'pointer',
    color: '#475569',
    marginLeft: '10px'
  },
  resetButton: {
    background: '#e2e8f0',
    border: 'none',
    padding: '8px 15px',
    borderRadius: '4px',
    cursor: 'pointer',
    color: '#475569'
  },
  error: {
    background: '#fee2e2',
    color: '#dc2626',
    padding: '12px',
    borderRadius: '4px',
    marginBottom: '20px',
    border: '1px solid #fecaca'
  },
  tabs: {
    display: 'flex',
    gap: '10px',
    marginBottom: '30px'
  },
  tab: {
    flex: 1,
    padding: '12px',
    background: '#f8fafc',
    border: '1px solid #e2e8f0',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '16px',
    color: '#475569'
  },
  activeTab: {
    flex: 1,
    padding: '12px',
    background: '#1a365d',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '16px',
    fontWeight: 'bold'
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '20px',
    background: 'white',
    padding: '30px',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
  },
  formGroup: {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px'
  },
  label: {
    fontWeight: '600',
    color: '#1e293b'
  },
  input: {
    padding: '12px',
    border: '1px solid #e2e8f0',
    borderRadius: '4px',
    fontSize: '16px',
    transition: 'border-color 0.2s',
    outline: 'none'
  },
  select: {
    padding: '12px',
    border: '1px solid #e2e8f0',
    borderRadius: '4px',
    fontSize: '16px',
    background: 'white',
    cursor: 'pointer'
  },
  fileInput: {
    padding: '10px',
    border: '1px dashed #cbd5e1',
    borderRadius: '4px',
    background: '#f8fafc',
    cursor: 'pointer'
  },
  fileName: {
    color: '#64748b',
    fontSize: '14px',
    marginTop: '4px'
  },
  buttonGroup: {
    marginTop: '20px'
  },
  submitButton: {
    width: '100%',
    background: '#1a365d',
    color: 'white',
    border: 'none',
    padding: '16px',
    borderRadius: '6px',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
    transition: 'background 0.2s',
    ':hover': {
      background: '#2d4b7a'
    },
    ':disabled': {
      background: '#94a3b8',
      cursor: 'not-allowed'
    }
  },
  tips: {
    marginTop: '30px',
    padding: '20px',
    background: '#f0f9ff',
    borderRadius: '8px',
    border: '1px solid #bae6fd'
  },
  tipsList: {
    margin: '10px 0 0 20px',
    color: '#0369a1',
    lineHeight: '1.6'
  }
};

// Make sure inputs get focus styles
styles.input[':focus'] = {
  borderColor: '#1a365d',
  boxShadow: '0 0 0 2px rgba(26, 54, 93, 0.1)'
};
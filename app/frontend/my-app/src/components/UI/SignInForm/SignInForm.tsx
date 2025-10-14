import React, { useState, useEffect } from 'react';
import { useAuth } from '../../../context/AuthContext';
import "./SignInForm.scss";

interface SignInFormProps {
  onSuccess?: () => void;
}

const SignInForm: React.FC<SignInFormProps> = ({ onSuccess }) => {
  const { login } = useAuth();
  const [formData, setFormData] = useState({
    login: '',
    password: ''
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showError, setShowError] = useState(false);

  // Эффект для анимации появления ошибки
  useEffect(() => {
    if (error) {
      setShowError(true);
    } else {
      setShowError(false);
    }
  }, [error]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      await login(formData.login, formData.password);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка авторизации');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className='signIn'>
      <form className='signInForm' onSubmit={handleSubmit}>
        <div className='signInForm__field'>
          <input
            type="text"
            name="login"
            placeholder="Логин"
            value={formData.login}
            onChange={handleChange}
            required
            className='signInForm__input'
          />
        </div>

        <div className='signInForm__field'>
          <input
            type="password"
            name="password"
            placeholder="Пароль"
            value={formData.password}
            onChange={handleChange}
            required
            className='signInForm__input'
          />
        </div>

        <div className={`signInForm__error ${showError ? 'signInForm__error--visible' : ''}`}>
          {error}
        </div>

        <div className={`signInForm__field ${showError ? 'signInForm__field--with-error' : ''}`}>
          <button 
            type="submit" 
            className='signInForm__button'
            disabled={loading}
          >
            {loading ? 'Вход...' : 'Вход'}
          </button>
        </div>
      </form>
      
      <div className="signInWarning">
        <p className='signInWarning__title'>
          При входе вы принимаете <a href='/' className='signInWarning__title--underlined'>условия использования</a> сайта и соглашаетесь на обработку 
          персональных данных согласно <a href='/' className='signInWarning__title--underlined'>политике конфиденциальности.</a>
        </p>
      </div>
    </div>
  );
};

export default SignInForm;

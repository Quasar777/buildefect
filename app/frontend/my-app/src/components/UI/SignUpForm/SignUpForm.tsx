import React, { useState, useEffect } from 'react';
import { useAuth } from '../../../context/AuthContext';
import "./SignUpForm.scss";

interface SignUpFormProps {
  onSuccess?: () => void;
}

const SignUpForm: React.FC<SignUpFormProps> = ({ onSuccess }) => {
  const { register } = useAuth();
  const [formData, setFormData] = useState({
    login: '',
    password: '',
    confirmPassword: '',
    name: '',
    lastname: ''
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

    // Валидация
    if (formData.password !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      setLoading(false);
      return;
    }

    if (formData.password.length < 6) {
      setError('Пароль должен содержать минимум 6 символов');
      setLoading(false);
      return;
    }

    try {
      await register(formData.login, formData.password, formData.name, formData.lastname);
      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка регистрации');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className='signUp'>
      <form className='signUpForm' onSubmit={handleSubmit}>
        <div className='signUpForm__field'>
          <input
            type="text"
            name="login"
            placeholder="Логин"
            value={formData.login}
            onChange={handleChange}
            required
            className='signUpForm__input'
          />
        </div>

        <div className='signUpForm__field'>
          <input
            type="text"
            name="name"
            placeholder="Имя"
            value={formData.name}
            onChange={handleChange}
            required
            className='signUpForm__input'
          />
        </div>

        <div className='signUpForm__field'>
          <input
            type="text"
            name="lastname"
            placeholder="Фамилия"
            value={formData.lastname}
            onChange={handleChange}
            required
            className='signUpForm__input'
          />
        </div>

        <div className='signUpForm__field'>
          <input
            type="password"
            name="password"
            placeholder="Пароль"
            value={formData.password}
            onChange={handleChange}
            required
            className='signUpForm__input'
          />
        </div>

        <div className='signUpForm__field'>
          <input
            type="password"
            name="confirmPassword"
            placeholder="Повторите пароль"
            value={formData.confirmPassword}
            onChange={handleChange}
            required
            className='signUpForm__input'
          />
        </div>

        <div className={`signUpForm__error ${showError ? 'signUpForm__error--visible' : ''}`}>
          {error}
        </div>

        <div className={`signUpForm__field ${showError ? 'signUpForm__field--with-error' : ''}`}>
          <button 
            type="submit" 
            className='signUpForm__button'
            disabled={loading}
          >
            {loading ? 'Регистрация...' : 'Регистрация'}
          </button>
        </div>
      </form>
      
      <div className="signUpWarning">
        <p className='signUpWarning__title'>
          При регистрации вы принимаете <a href='/' className='signUpWarning__title--underlined'>условия использования</a> сайта и соглашаетесь на обработку 
          персональных данных согласно <a href='/' className='signUpWarning__title--underlined'>политике конфиденциальности.</a>
        </p>
      </div>
    </div>
  );
};

export default SignUpForm;

import React from 'react';
import { Link } from 'react-router-dom';
import SignInForm from '../../components/UI/SignInForm/SignInForm';
import { useNavigate } from 'react-router-dom';
import './SignInPage.scss';

const SignInPage: React.FC = () => {
  const navigate = useNavigate();

  const handleSuccess = () => {
    navigate('/');
  };

  return (
    <div className='signInPage'>
      <div className='signInPage__container'>
        <h1 className='signInPage__title'>Вход в систему</h1>
        <SignInForm onSuccess={handleSuccess} />
        <div className='signInPage__footer'>
          <p className='signInPage__footer-text'>
            Нет аккаунта? <Link to="/signup" className='signInPage__footer-link'>Зарегистрироваться</Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default SignInPage;

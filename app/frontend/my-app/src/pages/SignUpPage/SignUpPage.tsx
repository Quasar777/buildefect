import React from 'react';
import { Link } from 'react-router-dom';
import SignUpForm from '../../components/UI/SignUpForm/SignUpForm';
import { useNavigate } from 'react-router-dom';
import './SignUpPage.scss';

const SignUpPage: React.FC = () => {
  const navigate = useNavigate();

  const handleSuccess = () => {
    navigate('/');
  };

  return (
    <div className='signUpPage'>
      <div className='signUpPage__container'>
        <h1 className='signUpPage__title'>Регистрация</h1>
        <SignUpForm onSuccess={handleSuccess} />
        <div className='signUpPage__footer'>
          <p className='signUpPage__footer-text'>
            Уже есть аккаунт? <Link to="/signin" className='signUpPage__footer-link'>Войти</Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default SignUpPage;

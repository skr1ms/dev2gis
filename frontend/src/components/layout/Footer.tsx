import React from 'react';

export function Footer() {
  return (
    <footer className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 mt-auto transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div>
            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">О системе</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Автоматизированная система обработки данных БПЛА для генерации
              топографических карт высот с использованием фотограмметрии.
            </p>
          </div>
          
          <div>
            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Контакты</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Для вопросов и предложений<br />
              свяжитесь с администрацией системы
            </p>
          </div>
        </div>
        
        <div className="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 text-center">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            © {new Date().getFullYear()} Карты высот БПЛА. Все права защищены.
          </p>
        </div>
      </div>
    </footer>
  );
}


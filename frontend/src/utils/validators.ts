export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

export function validateEmail(email: string): string | null {
  if (!email) return 'Email обязателен';
  if (!isValidEmail(email)) return 'Некорректный email адрес';
  return null;
}

export function validatePassword(password: string): string | null {
  if (!password) return 'Пароль обязателен';
  if (password.length < 8) return 'Пароль должен содержать минимум 8 символов';
  return null;
}

export function validatePasswordStrength(password: string): {
  score: number;
  feedback: string;
} {
  let score = 0;
  
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[a-z]/.test(password)) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[0-9]/.test(password)) score++;
  if (/[^a-zA-Z0-9]/.test(password)) score++;
  
  const feedback = 
    score < 2 ? 'Слабый' :
    score < 4 ? 'Средний' :
    score < 6 ? 'Хороший' : 'Отличный';
  
  return { score, feedback };
}

export function validateName(name: string): string | null {
  if (!name) return 'Имя обязательно';
  if (name.length < 2) return 'Имя должно содержать минимум 2 символа';
  return null;
}

export function validateFile(file: File, allowedTypes: string[], maxSizeMB: number = 100): string | null {
  if (!file) return 'Файл не выбран';
  
  const fileType = file.type.toLowerCase();
  const isValidType = allowedTypes.some(type => fileType.includes(type));
  
  if (!isValidType) {
    return `Неподдерживаемый формат. Разрешены: ${allowedTypes.join(', ')}`;
  }
  
  const fileSizeMB = file.size / (1024 * 1024);
  if (fileSizeMB > maxSizeMB) {
    return `Файл слишком большой. Максимум: ${maxSizeMB}МБ`;
  }
  
  return null;
}

export function validateImageFiles(files: File[]): string | null {
  if (files.length === 0) return 'Выберите хотя бы один файл';
  
  const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png'];
  
  for (const file of files) {
    const error = validateFile(file, allowedTypes);
    if (error) return error;
  }
  
  return null;
}

export function validateBatchFiles(files: File[]): string | null {
  if (files.length < 5) return 'Для пакетной обработки требуется минимум 5 изображений';
  if (files.length > 50) return 'Максимум 50 изображений для пакетной обработки';
  
  return validateImageFiles(files);
}


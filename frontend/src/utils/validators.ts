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
  const fileName = file.name.toLowerCase();
  const isValidType = allowedTypes.some(type => {
    const cleanType = type.replace('image/', '').replace('/', '');
    return fileType.includes(cleanType) || fileName.endsWith(`.${cleanType}`);
  });
  
  if (!isValidType) {
    const allowedExtensions = allowedTypes.map(t => t.replace('image/', '')).join(', ');
    return `Неподдерживаемый формат файла "${file.name}". Разрешены: ${allowedExtensions}`;
  }
  
  const fileSizeMB = file.size / (1024 * 1024);
  if (fileSizeMB > maxSizeMB) {
    return `Файл "${file.name}" слишком большой (${fileSizeMB.toFixed(2)}МБ). Максимум: ${maxSizeMB}МБ`;
  }
  
  if (file.size === 0) {
    return `Файл "${file.name}" пустой или поврежден`;
  }
  
  return null;
}

export function validateImageFiles(files: File[]): string | null {
  if (files.length === 0) return 'Выберите хотя бы один файл';
  
  if (files.length > 50) return 'Максимум 50 файлов за один раз';
  
  const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'jpeg', 'jpg', 'png'];
  
  for (const file of files) {
    const error = validateFile(file, allowedTypes, 50);
    if (error) return error;
  }
  
  // Check for duplicate filenames
  const fileNames = files.map(f => f.name);
  const duplicates = fileNames.filter((name, index) => fileNames.indexOf(name) !== index);
  if (duplicates.length > 0) {
    return `Обнаружены дублирующиеся имена файлов: ${duplicates.join(', ')}`;
  }
  
  return null;
}

export function validateBatchFiles(files: File[]): string | null {
  if (files.length < 5) return 'Для пакетной обработки требуется минимум 5 изображений';
  if (files.length > 50) return 'Максимум 50 изображений для пакетной обработки';
  
  const imageValidation = validateImageFiles(files);
  if (imageValidation) return imageValidation;
  
  // Calculate total size
  const totalSizeMB = files.reduce((sum, file) => sum + file.size, 0) / (1024 * 1024);
  if (totalSizeMB > 500) {
    return `Общий размер файлов слишком большой (${totalSizeMB.toFixed(2)}МБ). Максимум: 500МБ`;
  }
  
  return null;
}


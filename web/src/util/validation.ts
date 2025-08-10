import { ref } from 'vue'

/**
 * 表单校验工具类
 */

// 类型定义
export interface ValidationResult {
  isValid: boolean
  message: string
}

export interface FormValidationResult {
  isValid: boolean
  errors: Record<string, string>
}

export interface InputConfig {
  col: string
  label?: string
  subLabel?: string
  type?: string
  required?: boolean
  minLength?: number
  maxLength?: number
}

/**
 * 校验单个字段
 * @param fieldName 字段名
 * @param value 字段值
 * @param config 字段配置
 * @returns 校验结果
 */
export function validateField(fieldName: string, value: any, config: InputConfig): ValidationResult {
  const fieldLabel = config.subLabel || config.label || fieldName
  
  // 基本的非空校验
  if (config.required !== false && (!value || value.toString().trim() === '')) {
    return {
      isValid: false,
      message: `${fieldLabel}不能为空`
    }
  }
  
  // 如果值为空且不是必填，则跳过其他校验
  if (!value || value.toString().trim() === '') {
    return { isValid: true, message: '' }
  }
  
  // 邮箱校验
  if (config.type === 'email') {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(value)) {
      return {
        isValid: false,
        message: '请输入有效的邮箱地址'
      }
    }
  }
  
  // 手机号校验
  if (config.type === 'phone') {
    const phoneRegex = /^1[3-9]\d{9}$/
    if (!phoneRegex.test(value)) {
      return {
        isValid: false,
        message: '请输入有效的手机号码'
      }
    }
  }
  
  // URL校验
  if (config.type === 'url') {
    try {
      new URL(value)
    } catch {
      return {
        isValid: false,
        message: '请输入有效的URL地址'
      }
    }
  }
  
  // 数字校验
  if (config.type === 'number') {
    if (isNaN(Number(value))) {
      return {
        isValid: false,
        message: '请输入有效的数字'
      }
    }
  }
  
  // 长度校验
  const valueStr = value.toString()
  if (config.minLength && valueStr.length < config.minLength) {
    return {
      isValid: false,
      message: `最少需要${config.minLength}个字符`
    }
  }
  
  if (config.maxLength && valueStr.length > config.maxLength) {
    return {
      isValid: false,
      message: `最多允许${config.maxLength}个字符`
    }
  }
  
  return { isValid: true, message: '' }
}

/**
 * 校验整个表单
 * @param formData 表单数据
 * @param inputConfigs 输入字段配置数组
 * @returns 校验结果
 */
export function validateForm(formData: Record<string, any>, inputConfigs: InputConfig[]): FormValidationResult {
  const errors: Record<string, string> = {}
  
  inputConfigs.forEach(config => {
    const value = formData[config.col]
    const result = validateField(config.col, value, config)
    
    if (!result.isValid) {
      errors[config.col] = result.message
    }
  })
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors
  }
}

/**
 * 创建响应式校验状态管理
 * @param initialErrors 初始错误状态
 * @returns 校验状态管理对象
 */
export function createValidationState(initialErrors: Record<string, string> = {}) {
  const errors = ref(initialErrors)
  
  return {
    errors,
    
    // 设置字段错误
    setFieldError(fieldName: string, message: string) {
      errors.value[fieldName] = message
    },
    
    // 清除字段错误
    clearFieldError(fieldName: string) {
      delete errors.value[fieldName]
    },
    
    // 清除所有错误
    clearAllErrors() {
      errors.value = {}
    },
    
    // 设置多个错误
    setErrors(newErrors: Record<string, string>) {
      errors.value = { ...newErrors }
    },
    
    // 检查是否有错误
    hasErrors() {
      return Object.keys(errors.value).length > 0
    },
    
    // 获取字段错误
    getFieldError(fieldName: string) {
      return errors.value[fieldName] || ''
    }
  }
}

/**
 * 检查表单是否有空值字段
 * @param formData 表单数据
 * @param inputConfigs 输入字段配置数组
 * @returns 是否有空值字段
 */
export function hasEmptyRequiredFields(formData: Record<string, any>, inputConfigs: InputConfig[]): boolean {
  return inputConfigs.some(config => {
    if (config.required === false) return false
    const value = formData[config.col]
    return !value || value.toString().trim() === ''
  })
}
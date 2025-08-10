import { DefineComponent } from 'vue';

declare const CardNum: DefineComponent<{
  title: string;
  value: string | number;
  description?: string;
  icon?: any;
}>;

export default CardNum;
import useClipboard from "vue-clipboard3";
import { ElMessage } from 'element-plus'

function copyToClipboard(obj) {
    const { toClipboard } = useClipboard();
    try {
        if (!obj) {
            ElMessage({
                message: `复制内容为空`,
                type: 'error',
            })
        } else {
            toClipboard(obj);
            ElMessage({
                message: '复制成功',
                type: 'success',
            })
        }
    } catch (e) {
        ElMessage({
            message: `复制失败，${e}`,
            type: 'error',
        })
    }
};

export { copyToClipboard };

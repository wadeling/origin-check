export default function AboutPage() {
  return (
    <article className="prose prose-slate max-w-none">
      <h1>评测方法论</h1>
      <p>
        Origin Check 通过平台统一调度，定期对收录的 AI API 中转站进行探测，从<strong>模型真伪</strong>与<strong>性能稳定性</strong>两个维度给出参考评分。
      </p>

      <h2>真伪鉴定</h2>
      <ul>
        <li><strong>Metadata 校验</strong>：对比宣称模型与响应中的 model 字段（权重 15%）</li>
        <li><strong>指纹 Prompt 套件</strong>：格式/逻辑/自报模型/推理/拒绝边界等用例（权重 40%）</li>
        <li><strong>能力边界</strong>：根据 prompt 合规性推断基础能力（权重 25%）</li>
        <li><strong>Cache 重复探测</strong>：相同 prompt 连发 3 次，检测 CDN/语义缓存（权重 20%）</li>
      </ul>
      <p>最终输出 0–100 一致性得分，并标注「一致 / 存疑 / 不符」。</p>

      <h2>性能评测</h2>
      <ul>
        <li><strong>TTFT</strong>：流式请求首 token 时间</li>
        <li><strong>E2E 延迟</strong>：非流式完整响应时间</li>
        <li><strong>24h 可用率</strong>：health 探测成功率（详情页卡片）</li>
        <li><strong>近 7 天性能趋势</strong>：每日 performance 探测汇总</li>
      </ul>

      <h2>调度策略</h2>
      <p>间隔通过环境变量配置（Go duration 格式）：</p>
      <ul>
        <li>Health 探测：<code>PROBE_HEALTH_INTERVAL</code>，默认 15 分钟</li>
        <li>Performance 探测：<code>PROBE_PERFORMANCE_INTERVAL</code>，默认 24 小时（每天一次）</li>
        <li>Authenticity 鉴定：<code>PROBE_AUTHENTICITY_INTERVAL</code>，默认 24 小时（一天一次）</li>
      </ul>

      <p className="text-sm text-muted">评测结果仅供参考，不构成商业指控。</p>
    </article>
  );
}
